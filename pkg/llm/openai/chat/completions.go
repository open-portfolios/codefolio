package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/open-portfolios/codefolio/pkg/llm"
	"github.com/open-portfolios/codefolio/pkg/stdx"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
)

var (
	_ llm.Driver = (*CompletionsDriver)(nil)
)

type CompletionsDriver struct {
	client *openai.Client
}

func NewCompletionsDriver(client *openai.Client) *CompletionsDriver {
	return &CompletionsDriver{
		client: client,
	}
}

func (c *CompletionsDriver) Stream(ctx context.Context, messages []llm.Message, options ...llm.StreamOption) (<-chan llm.Delta, <-chan error) {
	conf := llm.CollectStreamOptions(options...)

	deltaChan := make(chan llm.Delta, conf.ChanCapacity)
	errChan := make(chan error, 1)
	go func() {
		defer close(deltaChan)
		defer close(errChan)

		converter := &messageConverter{}
		for _, msg := range messages {
			if err := msg.Accept(converter); err != nil {
				errChan <- err
				return
			}
		}

		tools, err := buildTools(conf.Tools)
		if err != nil {
			errChan <- err
			return
		}

		stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    conf.Model,
			Messages: converter.msgs,
			Tools:    tools,
		})
		defer stream.Close()

		for stream.Next() {
			event := stream.Current()

			if reasoning := extractReasoning(event); reasoning != "" {
				t := llm.ThinkingDelta{
					Content: reasoning,
				}
				if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, t); err != nil {
					errChan <- err
					return
				}
			}

			if len(event.Choices) > 0 {
				m := llm.MessageDelta{
					Role:    event.Choices[0].Delta.Role,
					Content: event.Choices[0].Delta.Content,
				}
				if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, m); err != nil {
					errChan <- err
					return
				}
			}

			if len(event.Choices) > 0 {
				for _, tc := range event.Choices[0].Delta.ToolCalls {
					if tc.ID != "" {
						start := llm.ToolCallStartDelta{
							Index: int(tc.Index),
							ID:    tc.ID,
							Name:  tc.Function.Name,
						}
						if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, start); err != nil {
							errChan <- err
							return
						}
					}
					if tc.Function.Arguments != "" {
						input := llm.ToolCallInputDelta{
							Index: int(tc.Index),
							Input: tc.Function.Arguments,
						}
						if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, input); err != nil {
							errChan <- err
							return
						}
					}
				}
			}

			if event.Usage.TotalTokens > 0 {
				u := llm.UsageDelta{
					InputTokens:  uint64(event.Usage.PromptTokens),
					OutputTokens: uint64(event.Usage.CompletionTokens),
					TotalTokens:  uint64(event.Usage.TotalTokens),
					Final:        true,
				}
				if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, u); err != nil {
					errChan <- err
					return
				}
			}

			if len(event.Choices) > 0 && event.Choices[0].FinishReason != "" {
				stopReason := "end_turn"
				if event.Choices[0].FinishReason == "tool_calls" {
					stopReason = "tool_use"
				}
				ss := llm.StreamStopDelta{
					StopReason: stopReason,
				}
				if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, ss); err != nil {
					errChan <- err
					return
				}
			}
		}
		if err := stream.Err(); err != nil {
			errChan <- fmt.Errorf("openai stream error: %w", err)
			return
		}
	}()
	return deltaChan, errChan
}

func buildTools(schemas []map[string]any) ([]openai.ChatCompletionToolUnionParam, error) {
	tools := make([]openai.ChatCompletionToolUnionParam, 0, len(schemas))
	for i, schema := range schemas {
		name, ok := schema["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("tool schema %d has no name", i)
		}
		desc, _ := schema["description"].(string)
		inputSchema, ok := schema["input_schema"].(map[string]any)
		if !ok || inputSchema == nil {
			return nil, fmt.Errorf("tool schema %q has no input_schema", name)
		}
		tools = append(tools, openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: openai.FunctionDefinitionParam{
					Name:        name,
					Description: param.NewOpt(desc),
					Parameters:  openai.FunctionParameters(inputSchema),
				},
			},
		})
	}
	return tools, nil
}

func extractReasoning(chunk openai.ChatCompletionChunk) string {
	for _, choice := range chunk.Choices {
		rc := choice.Delta.JSON.ExtraFields["reasoning_content"]
		if !rc.Valid() {
			continue
		}
		var v string
		if err := json.Unmarshal([]byte(rc.Raw()), &v); err == nil {
			return v
		}
	}
	return ""
}
