package messages

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
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
		defer close(errChan)
		defer close(deltaChan)

		sys := make([]anthropic.TextBlockParam, 0)
		msgs := make([]anthropic.MessageParam, 0)
		for _, msg := range messages {
			switch msg.Role() {
			case llm.RoleSystem:
				sys = append(sys, anthropic.TextBlockParam{Text: msg.Content()})
			case llm.RoleUser:
				msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content())))
			case llm.RoleAssistant:
				msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content())))
			default:
				errChan <- fmt.Errorf("unsupported role: %s", msg.Role())
				return
			}
		}

		stream := d.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
			Model:     conf.Model,
			MaxTokens: conf.MaxTokens,
			System:    sys,
			Messages:  msgs,
		})
		defer stream.Close()

		for stream.Next() {
			event := stream.Current()
			switch e := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				m := llm.MessageDelta{
					Role:    llm.RoleAssistant,
					Content: e.Delta.Text,
				}
				if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, m); err != nil {
					errChan <- err
					return
				}
			case anthropic.MessageDeltaEvent:
				u := llm.UsageDelta{
					TotalTokens: uint64(e.Usage.InputTokens) + uint64(e.Usage.OutputTokens),
				}
				if err := stdx.CancellableSend[llm.Delta](ctx, deltaChan, u); err != nil {
					errChan <- err
					return
				}
			}
		}
		if err := stream.Err(); err != nil {
			errChan <- err
			return
		}
	}()
	return deltaChan, errChan
}
