package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/open-portfolios/codefolio/pkg/llm"
)

type streamDeltaMsg string
type streamThinkingMsg string
type streamDoneMsg struct{}
type streamErrMsg struct{ Err error }
type spinnerTickMsg struct{}

func StreamLLM(p *tea.Program, driver llm.Driver, messages []llm.Message, model string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		deltaCh, errCh := driver.Stream(ctx, messages, llm.WithModel(model))

		go func() {
			for {
				select {
				case delta, ok := <-deltaCh:
					if !ok {
						p.Send(streamDoneMsg{})
						return
					}

					c := &collector{}
					if err := delta.Accept(c); err != nil {
						p.Send(streamErrMsg{Err: err})
						return
					}
					if c.thinking != "" {
						p.Send(streamThinkingMsg(c.thinking))
					}
					if c.message != "" {
						p.Send(streamDeltaMsg(c.message))
					}

				case err, ok := <-errCh:
					if ok && err != nil {
						p.Send(streamErrMsg{Err: err})
						return
					}
				}
			}
		}()

		return nil
	}
}

type collector struct {
	llm.BaseDeltaVisitor
	message  string
	thinking string
}

func (c *collector) VisitMessage(d llm.MessageDelta) error {
	c.message = d.Content
	return nil
}

func (c *collector) VisitThinking(d llm.ThinkingDelta) error {
	c.thinking = d.Content
	return nil
}
