package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/open-portfolios/codefolio/pkg/llm"
)

type streamDeltaMsg string
type streamDoneMsg struct{}
type streamErrMsg struct{ Err error }

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

					var msgDelta llm.MessageDelta
					if err := delta.Accept(&collector{delta: &msgDelta}); err != nil {
						p.Send(streamErrMsg{Err: err})
						return
					}

					if msgDelta.Content != "" {
						p.Send(streamDeltaMsg(msgDelta.Content))
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
	delta *llm.MessageDelta
}

func (c *collector) VisitMessage(d llm.MessageDelta) error {
	*c.delta = d
	return nil
}
