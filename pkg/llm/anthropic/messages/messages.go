package messages

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/open-portfolios/codefolio/pkg/llm"
	"github.com/open-portfolios/codefolio/pkg/stdx"
)

var (
	_ llm.Driver = (*Driver)(nil)
)

type Driver struct {
	client *anthropic.Client
}

func NewDriver(client *anthropic.Client) *Driver {
	return &Driver{
		client: client,
	}
}

func (d *Driver) Stream(ctx context.Context, messages []llm.Message, options ...llm.StreamOption) (<-chan llm.Delta, <-chan error) {
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
		converter.flushToolBlocks()

		tools := make([]anthropic.ToolUnionParam, 0, len(conf.Tools))
		for _, schema := range conf.Tools {
			name, _ := schema["name"].(string)
			desc, _ := schema["description"].(string)
			inputSchema, _ := schema["input_schema"].(map[string]any)
			var props any
			var required []string
			if inputSchema != nil {
				if p, ok := inputSchema["properties"]; ok {
					props = p
				}
				if r, ok := inputSchema["required"]; ok {
					if rs, ok := r.([]string); ok {
						required = rs
					} else if rs, ok := r.([]any); ok {
						for _, item := range rs {
							if s, ok := item.(string); ok {
								required = append(required, s)
							}
						}
					}
				}
			}
			tools = append(tools, anthropic.ToolUnionParam{
				OfTool: &anthropic.ToolParam{
					Name:        name,
					Description: param.NewOpt(desc),
					InputSchema: anthropic.ToolInputSchemaParam{
						Properties: props,
						Required:   required,
					},
				},
			})
		}

		stream := d.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
			Model:     conf.Model,
			MaxTokens: conf.MaxTokens,
			System:    converter.sys,
			Messages:  converter.msgs,
			Tools:     tools,
		})
		defer stream.Close()

		toolCallIdx := 0
		hasToolUse := false
		for stream.Next() {
			event := stream.Current()
			switch e := event.AsAny().(type) {
			case anthropic.ContentBlockStartEvent:
				if e.ContentBlock.Type == "thinking" {
					ts := llm.ThinkingStartDelta{
						Signature: e.ContentBlock.Signature,
					}
					if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, ts); err != nil {
						errChan <- err
						return
					}
				}
				if e.ContentBlock.Type == "tool_use" {
					tc := llm.ToolCallStartDelta{
						Index: toolCallIdx,
						ID:    e.ContentBlock.ID,
						Name:  e.ContentBlock.Name,
					}
					if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, tc); err != nil {
						errChan <- err
						return
					}
					toolCallIdx++
					hasToolUse = true
				}
			case anthropic.ContentBlockDeltaEvent:
				switch delta := e.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					m := llm.MessageDelta{
						Role:    llm.RoleAssistant,
						Content: delta.Text,
					}
					if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, m); err != nil {
						errChan <- err
						return
					}
				case anthropic.ThinkingDelta:
					t := llm.ThinkingDelta{
						Content: delta.Thinking,
					}
					if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, t); err != nil {
						errChan <- err
						return
					}
				case anthropic.InputJSONDelta:
					ti := llm.ToolCallInputDelta{
						Index: toolCallIdx - 1,
						Input: delta.PartialJSON,
					}
					if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, ti); err != nil {
						errChan <- err
						return
					}
				}
			case anthropic.MessageDeltaEvent:
				u := llm.UsageDelta{
					InputTokens:  uint64(e.Usage.InputTokens),
					OutputTokens: uint64(e.Usage.OutputTokens),
					TotalTokens:  uint64(e.Usage.InputTokens) + uint64(e.Usage.OutputTokens),
				}
				if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, u); err != nil {
					errChan <- err
					return
				}
			}
		}
		if err := stream.Err(); err != nil {
			errChan <- fmt.Errorf("anthropic stream error: %w", err)
			return
		}

		stopReason := "end_turn"
		if hasToolUse {
			stopReason = "tool_use"
		}
		ss := llm.StreamStopDelta{StopReason: stopReason}
		if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, ss); err != nil {
			errChan <- err
			return
		}
	}()
	return deltaChan, errChan
}
