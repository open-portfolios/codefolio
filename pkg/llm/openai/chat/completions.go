package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/open-portfolios/codefolio/pkg/llm"
	"github.com/open-portfolios/codefolio/pkg/stdx"
	"github.com/openai/openai-go/v3"
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
		defer close(errChan)
		defer close(deltaChan)

		msgs := make([]openai.ChatCompletionMessageParamUnion, 0)
		for _, msg := range messages {
			switch msg.Role() {
			case llm.RoleUser:
				msgs = append(msgs, openai.UserMessage(msg.Content()))
			case llm.RoleAssistant:
				msgs = append(msgs, openai.AssistantMessage(msg.Content()))
			case llm.RoleDeveloper:
				msgs = append(msgs, openai.DeveloperMessage(msg.Content()))
			case llm.RoleSystem:
				msgs = append(msgs, openai.SystemMessage(msg.Content()))
			case llm.RoleTool:
				if toolcall, ok := msg.(llm.ToolCallMessage); ok {
					msgs = append(msgs, openai.ToolMessage(toolcall.Content(), toolcall.ToolCallID()))
				} else {
					errChan <- llm.ErrMalformedToolMessage
					return
				}
			case llm.RoleFunction:
				if function, ok := msg.(llm.FunctionMessage); ok {
					msgs = append(msgs, openai.ChatCompletionMessageParamOfFunction(function.Content(), function.Name()))
				} else {
					errChan <- llm.ErrMalformedFunctionMessage
					return
				}
			default:
				errChan <- fmt.Errorf("unsupported role: %s", msg.Role())
				return
			}
		}

		stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    conf.Model,
			Messages: msgs,
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

			m := llm.MessageDelta{
				Role:    event.Choices[0].Delta.Role,
				Content: event.Choices[0].Delta.Content,
			}
			if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, m); err != nil {
				errChan <- err
				return
			}

			u := llm.UsageDelta{
				TotalTokens: uint64(event.Usage.TotalTokens),
			}
			if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, u); err != nil {
				errChan <- err
				return
			}
		}
		if err := stream.Err(); err != nil {
			errChan <- err
			return
		}
	}()
	return deltaChan, errChan
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
