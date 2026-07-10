package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools"
	"github.com/open-portfolios/codefolio/internal/svc"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

type streamDeltaMsg string
type streamThinkingMsg string
type thinkingStartMsg struct{ Signature string }
type streamDoneMsg struct{}
type streamErrMsg struct{ Err error }
type spinnerTickMsg struct{}
type toolCallMsg struct {
	ID       string
	ToolName string
	Input    string
	Content  string
	IsDone   bool
}

func RunAgent(ctx context.Context, p *tea.Program, agent *svc.Agent, driver llm.Driver, session *domain.Session, registry *tools.Registry, model string) tea.Cmd {
	return func() tea.Msg {
		err := agent.Run(ctx, driver, session, registry, model, func(event svc.Event) {
			switch event.Type {
			case svc.EventDelta:
				p.Send(streamDeltaMsg(event.Content))
			case svc.EventThinking:
				p.Send(streamThinkingMsg(event.Content))
			case svc.EventThinkingStart:
				p.Send(thinkingStartMsg{Signature: event.Content})
			case svc.EventToolStart:
				p.Send(toolCallMsg{
					ID:       event.ToolID,
					ToolName: event.ToolName,
					Input:    event.ToolInput,
				})
			case svc.EventToolDone:
				p.Send(toolCallMsg{
					ID:       event.ToolID,
					ToolName: event.ToolName,
					Content:  event.Content,
					IsDone:   true,
				})
			case svc.EventDone:
				p.Send(streamDoneMsg{})
			case svc.EventError:
				p.Send(streamErrMsg{Err: event.Err})
			}
		})
		if err != nil {
			p.Send(streamErrMsg{Err: err})
		}
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
