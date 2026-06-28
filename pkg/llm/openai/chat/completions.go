package chat

import (
	"context"
	"fmt"

	"github.com/open-portfolios/codefolio/pkg/llm"
	"github.com/openai/openai-go/v3"
)

var (
	_ llm.Driver = (*CompletionsDriver)(nil)
)

type CompletionsDriver struct {
	client       *openai.Client
	chanCapacity int
}

type option func(*CompletionsDriver)

func NewCompletionsDriver(client *openai.Client, options ...option) *CompletionsDriver {
	c := &CompletionsDriver{
		client:       client,
		chanCapacity: 64,
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

func WithChanCapacity(capacity int) option {
	return func(c *CompletionsDriver) { c.chanCapacity = capacity }
}

func (c *CompletionsDriver) Stream(ctx context.Context, messages []llm.Message, options ...llm.StreamOption) (<-chan llm.Delta, <-chan error) {
	deltaChan := make(chan llm.Delta, c.chanCapacity)
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

		conf := llm.CollectStreamOptions(options...)
		stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    conf.Model,
			Messages: msgs,
		})
		defer stream.Close()

		for stream.Next() {
			event := stream.Current()
			delta := newCompletionsDelta(
				event.Choices[0].Delta.Role,
				event.Choices[0].Delta.Content,
				event.Usage.TotalTokens,
			)
			select {
			case deltaChan <- delta:
			case <-ctx.Done():
				errChan <- ctx.Err()
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
