package llm

import "github.com/open-portfolios/codefolio/pkg/stdx"

var (
	_ Delta = stdx.Zero[MessageDelta]()
	_ Delta = stdx.Zero[ThinkingDelta]()
	_ Delta = stdx.Zero[UsageDelta]()
)

type Delta interface {
	Accept(DeltaVisitor) error
}

type DeltaVisitor interface {
	VisitMessage(MessageDelta) error
	VisitThinking(ThinkingDelta) error
	VisitUsage(UsageDelta) error
}

type MessageDelta struct {
	Role    string
	Content string
}

type ThinkingDelta struct {
	Content string
}

type UsageDelta struct {
	TotalTokens uint64
}

func (m MessageDelta) Accept(d DeltaVisitor) error  { return d.VisitMessage(m) }
func (t ThinkingDelta) Accept(d DeltaVisitor) error  { return d.VisitThinking(t) }
func (u UsageDelta) Accept(d DeltaVisitor) error     { return d.VisitUsage(u) }

type BaseDeltaVisitor struct{}

func (BaseDeltaVisitor) VisitMessage(MessageDelta) error   { return nil }
func (BaseDeltaVisitor) VisitThinking(ThinkingDelta) error { return nil }
func (BaseDeltaVisitor) VisitUsage(UsageDelta) error       { return nil }
