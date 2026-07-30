package llm

import "github.com/open-portfolios/codefolio/pkg/stdx"

var (
	_ Delta = stdx.Zero[MessageDelta]()
	_ Delta = stdx.Zero[ThinkingDelta]()
	_ Delta = stdx.Zero[ThinkingStartDelta]()
	_ Delta = stdx.Zero[UsageDelta]()
	_ Delta = stdx.Zero[ToolCallStartDelta]()
	_ Delta = stdx.Zero[ToolCallInputDelta]()
	_ Delta = stdx.Zero[StreamStopDelta]()
)

type Delta interface {
	Accept(DeltaVisitor) error
}

type DeltaVisitor interface {
	VisitMessage(MessageDelta) error
	VisitThinking(ThinkingDelta) error
	VisitThinkingStart(ThinkingStartDelta) error
	VisitUsage(UsageDelta) error
	VisitToolCallStart(ToolCallStartDelta) error
	VisitToolCallInput(ToolCallInputDelta) error
	VisitStreamStop(StreamStopDelta) error
}

type MessageDelta struct {
	Role    string
	Content string
}

type ThinkingDelta struct {
	Content string
}

type ThinkingStartDelta struct {
	Signature string
}

type UsageDelta struct {
	InputTokens  uint64
	OutputTokens uint64
	TotalTokens  uint64
	Final        bool
}

type ToolCallStartDelta struct {
	Index int
	ID    string
	Name  string
}

type ToolCallInputDelta struct {
	Index int
	Input string
}

type StreamStopDelta struct {
	StopReason string
}

func (m MessageDelta) Accept(d DeltaVisitor) error        { return d.VisitMessage(m) }
func (t ThinkingDelta) Accept(d DeltaVisitor) error       { return d.VisitThinking(t) }
func (ts ThinkingStartDelta) Accept(d DeltaVisitor) error { return d.VisitThinkingStart(ts) }
func (u UsageDelta) Accept(d DeltaVisitor) error          { return d.VisitUsage(u) }
func (tc ToolCallStartDelta) Accept(d DeltaVisitor) error { return d.VisitToolCallStart(tc) }
func (ti ToolCallInputDelta) Accept(d DeltaVisitor) error { return d.VisitToolCallInput(ti) }
func (ss StreamStopDelta) Accept(d DeltaVisitor) error    { return d.VisitStreamStop(ss) }

type BaseDeltaVisitor struct{}

func (BaseDeltaVisitor) VisitMessage(MessageDelta) error             { return nil }
func (BaseDeltaVisitor) VisitThinking(ThinkingDelta) error           { return nil }
func (BaseDeltaVisitor) VisitThinkingStart(ThinkingStartDelta) error { return nil }
func (BaseDeltaVisitor) VisitUsage(UsageDelta) error                 { return nil }
func (BaseDeltaVisitor) VisitToolCallStart(ToolCallStartDelta) error { return nil }
func (BaseDeltaVisitor) VisitToolCallInput(ToolCallInputDelta) error { return nil }
func (BaseDeltaVisitor) VisitStreamStop(StreamStopDelta) error       { return nil }
