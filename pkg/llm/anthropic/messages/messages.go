package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

		sys := make([]anthropic.TextBlockParam, 0)
		msgs := make([]anthropic.MessageParam, 0)
		for i := 0; i < len(messages); i++ {
			msg := messages[i]
			switch msg.Role() {
			case llm.RoleSystem:
				sys = append(sys, anthropic.TextBlockParam{Text: msg.Content()})
			case llm.RoleUser:
				msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content())))
			case llm.RoleAssistant:
				var blocks []anthropic.ContentBlockParamUnion
				if msg.Content() != "" {
					blocks = append(blocks, anthropic.NewTextBlock(msg.Content()))
				}
				if mc, ok := msg.(llm.MessageWithToolCalls); ok && len(mc.ToolCalls()) > 0 {
					for _, tc := range mc.ToolCalls() {
						var input any
						if tc.Input != "" {
							json.Unmarshal([]byte(tc.Input), &input)
						}
						blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, input, tc.Name))
					}
				}
				if mt, ok := msg.(llm.MessageWithThinking); ok && mt.Thinking() != "" {
					blocks = append(blocks, anthropic.NewThinkingBlock(mt.ThinkingSignature(), mt.Thinking()))
				}
				if len(blocks) == 0 {
					blocks = append(blocks, anthropic.NewTextBlock(""))
				}
				msgs = append(msgs, anthropic.NewAssistantMessage(blocks...))
			case llm.RoleTool:
				var blocks []anthropic.ContentBlockParamUnion
				for i < len(messages) && messages[i].Role() == llm.RoleTool {
					tm := messages[i]
					if tc, ok := tm.(llm.ToolCallMessage); ok {
						isError := false
						if strings.HasPrefix(tc.Content(), "Error:") {
							isError = true
						}
						blocks = append(blocks, anthropic.NewToolResultBlock(tc.ToolCallID(), tc.Content(), isError))
					}
					i++
				}
				i--
				msgs = append(msgs, anthropic.NewUserMessage(blocks...))
			default:
				errChan <- fmt.Errorf("unsupported role: %s", msg.Role())
				return
			}
		}

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
			System:    sys,
			Messages:  msgs,
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
