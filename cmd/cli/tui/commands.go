package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/svc"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

type streamDeltaMsg string
type streamThinkingMsg string
type thinkingStartMsg struct{ Signature string }
type streamDoneMsg struct{}
type streamErrMsg struct{ Err error }
type spinnerTickMsg struct{}
type turnCompleteMsg struct{ Turn int }
type usageMsg struct {
	InputTokens  int64
	OutputTokens int64
}
type toolCallMsg struct {
	ID       string
	ToolName string
	Input    string
	Content  string
	IsDone   bool
}

func RunAgent(ctx context.Context, p *tea.Program, agent *svc.Agent, session domain.Session, cfg *conf.Global) tea.Cmd {
	return func() tea.Msg {
		var cb tuiEventVisitor
		cb.p = p
		err := agent.Run(ctx, session, cfg, &cb)
		if err != nil {
			p.Send(streamErrMsg{Err: err})
		}
		return nil
	}
}

type tuiEventVisitor struct {
	domain.BaseEventVisitor
	p *tea.Program
}

func (v *tuiEventVisitor) VisitStream(e domain.StreamEvent) error {
	v.p.Send(streamDeltaMsg(e.Content))
	return nil
}

func (v *tuiEventVisitor) VisitThink(e domain.ThinkEvent) error {
	v.p.Send(streamThinkingMsg(e.Content))
	return nil
}

func (v *tuiEventVisitor) VisitThinkStart(e domain.ThinkStartEvent) error {
	v.p.Send(thinkingStartMsg{Signature: e.Signature})
	return nil
}

func (v *tuiEventVisitor) VisitToolCall(e domain.ToolCallEvent) error {
	v.p.Send(toolCallMsg{
		ID:       e.ID,
		ToolName: e.Name,
		Input:    e.Input,
	})
	return nil
}

func (v *tuiEventVisitor) VisitToolResult(e domain.ToolResultEvent) error {
	v.p.Send(toolCallMsg{
		ID:       e.ID,
		ToolName: e.Name,
		Content:  e.Output,
		IsDone:   true,
	})
	return nil
}

func (v *tuiEventVisitor) VisitTurnComplete(e domain.TurnCompleteEvent) error {
	v.p.Send(turnCompleteMsg{Turn: e.Turn})
	return nil
}

func (v *tuiEventVisitor) VisitLoopComplete(e domain.LoopCompleteEvent) error {
	v.p.Send(streamDoneMsg{})
	return nil
}

func (v *tuiEventVisitor) VisitUsage(e domain.UsageEvent) error {
	v.p.Send(usageMsg{InputTokens: e.InputTokens, OutputTokens: e.OutputTokens})
	return nil
}

func (v *tuiEventVisitor) VisitError(e domain.ErrorEvent) error {
	v.p.Send(streamErrMsg{Err: e.Err})
	return nil
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
