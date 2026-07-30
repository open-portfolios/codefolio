package domain

import "time"

type EventVisitor interface {
	VisitStream(StreamEvent) error
	VisitThink(ThinkEvent) error
	VisitThinkStart(ThinkStartEvent) error
	VisitToolCall(ToolCallEvent) error
	VisitToolResult(ToolResultEvent) error
	VisitTurnComplete(TurnCompleteEvent) error
	VisitLoopComplete(LoopCompleteEvent) error
	VisitUsage(UsageEvent) error
	VisitError(ErrorEvent) error
}

type BaseEventVisitor struct{}

func (BaseEventVisitor) VisitStream(StreamEvent) error             { return nil }
func (BaseEventVisitor) VisitThink(ThinkEvent) error               { return nil }
func (BaseEventVisitor) VisitThinkStart(ThinkStartEvent) error     { return nil }
func (BaseEventVisitor) VisitToolCall(ToolCallEvent) error         { return nil }
func (BaseEventVisitor) VisitToolResult(ToolResultEvent) error     { return nil }
func (BaseEventVisitor) VisitTurnComplete(TurnCompleteEvent) error { return nil }
func (BaseEventVisitor) VisitLoopComplete(LoopCompleteEvent) error { return nil }
func (BaseEventVisitor) VisitUsage(UsageEvent) error               { return nil }
func (BaseEventVisitor) VisitError(ErrorEvent) error               { return nil }

type StreamEvent struct{ Content string }
type ThinkEvent struct{ Content string }
type ThinkStartEvent struct{ Signature string }

type ToolCallEvent struct {
	ID    string
	Name  string
	Input string
}

type ToolResultEvent struct {
	ID      string
	Name    string
	Output  string
	IsError bool
	Outcome ToolOutcome
	Elapsed time.Duration
}

type TurnCompleteEvent struct{ Turn int }
type LoopCompleteEvent struct{ TotalTurns int }
type UsageEvent struct{ InputTokens, OutputTokens int64 }
type ErrorEvent struct{ Err error }
