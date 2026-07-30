package controller

import (
	"context"

	"github.com/cylixlee/tux/app"
	"github.com/open-portfolios/codefolio/internal/domain"
)

type visitor struct {
	domain.BaseEventVisitor
	controller *Controller
	runtime    *app.App
	runID      uint64
	ctx        context.Context
}

func (v *visitor) post(fn func()) error {
	if v.runtime == nil {
		return app.ErrStopped
	}
	return v.runtime.Dispatch(v.ctx, fn)
}

func (v *visitor) active() bool {
	return v.runID == v.controller.runID && v.controller.streaming == StreamRunning
}

func (v *visitor) VisitStream(e domain.StreamEvent) error {
	return v.post(func() {
		if v.active() {
			if m := v.controller.currentAssistant(); m != nil {
				m.Content += e.Content
				v.controller.invalidate()
			}
		}
	})
}
func (v *visitor) VisitThink(e domain.ThinkEvent) error {
	return v.post(func() {
		if v.active() {
			if m := v.controller.currentAssistant(); m != nil {
				m.Thinking += e.Content
				v.controller.invalidate()
			}
		}
	})
}
func (v *visitor) VisitThinkStart(domain.ThinkStartEvent) error { return nil }
func (v *visitor) VisitToolCall(e domain.ToolCallEvent) error {
	return v.post(func() {
		if v.active() {
			_ = v.applyToolCall(e)
		}
	})
}

func (v *visitor) applyToolCall(e domain.ToolCallEvent) error {
	if m := v.controller.currentAssistant(); m != nil {
		for _, tool := range m.Tools {
			if tool.ID == e.ID {
				tool.Name, tool.Input = e.Name, e.Input
				v.controller.invalidate()
				return nil
			}
		}
		m.Tools = append(m.Tools, &Tool{ID: e.ID, Name: e.Name, Input: e.Input})
		v.controller.invalidate()
	}
	return nil
}
func (v *visitor) VisitToolResult(e domain.ToolResultEvent) error {
	return v.post(func() {
		if v.active() {
			if m := v.controller.currentAssistant(); m != nil {
				for _, tool := range m.Tools {
					if tool.ID == e.ID {
						tool.Done, tool.Output, tool.IsError, tool.Outcome, tool.Elapsed = true, e.Output, e.IsError, e.Outcome, e.Elapsed
						v.controller.invalidate()
						break
					}
				}
			}
		}
	})
}
func (v *visitor) VisitTurnComplete(domain.TurnCompleteEvent) error {
	return v.post(func() {
		if v.active() {
			v.controller.StartAssistantSegment()
		}
	})
}
func (v *visitor) VisitLoopComplete(domain.LoopCompleteEvent) error {
	return v.post(func() {
		if v.runID == v.controller.runID {
			v.controller.finish("")
			v.controller.invalidate()
		}
	})
}
func (v *visitor) VisitUsage(e domain.UsageEvent) error {
	return v.post(func() {
		if v.runID == v.controller.runID {
			v.controller.inputTokens, v.controller.outputTokens = e.InputTokens, e.OutputTokens
			v.controller.invalidate()
		}
	})
}
func (v *visitor) VisitError(e domain.ErrorEvent) error {
	return v.post(func() { v.controller.applyError(v.runID, e.Err) })
}
