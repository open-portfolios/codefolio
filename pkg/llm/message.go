package llm

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleDeveloper = "developer"
	RoleSystem    = "system"
	RoleTool      = "tool"
	RoleFunction  = "function"
)

var (
	_ MessageVisitor = (*BaseMessageVisitor)(nil)
)

type Message interface {
	Accept(MessageVisitor) error
}

type MessageVisitor interface {
	VisitUser(UserMessage) error
	VisitAssistant(AssistantMessage) error
	VisitSystem(SystemMessage) error
	VisitDeveloper(DeveloperMessage) error
	VisitTool(ToolMessage) error
	VisitFunction(FunctionMessage) error
}

type UserMessage struct {
	Content string
}

type AssistantMessage struct {
	Content           string
	Thinking          string
	ThinkingSignature string
	ToolCalls         []ToolCallInfo
}

type SystemMessage struct {
	Content string
}

type DeveloperMessage struct {
	Content string
}

type ToolMessage struct {
	Content    string
	ToolCallID string
}

type FunctionMessage struct {
	Content string
	Name    string
}

type ToolCallInfo struct {
	ID    string
	Name  string
	Input string
}

func (m UserMessage) Accept(v MessageVisitor) error      { return v.VisitUser(m) }
func (m AssistantMessage) Accept(v MessageVisitor) error { return v.VisitAssistant(m) }
func (m SystemMessage) Accept(v MessageVisitor) error    { return v.VisitSystem(m) }
func (m DeveloperMessage) Accept(v MessageVisitor) error { return v.VisitDeveloper(m) }
func (m ToolMessage) Accept(v MessageVisitor) error      { return v.VisitTool(m) }
func (m FunctionMessage) Accept(v MessageVisitor) error  { return v.VisitFunction(m) }

type BaseMessageVisitor struct{}

func (BaseMessageVisitor) VisitUser(UserMessage) error           { return nil }
func (BaseMessageVisitor) VisitAssistant(AssistantMessage) error { return nil }
func (BaseMessageVisitor) VisitSystem(SystemMessage) error       { return nil }
func (BaseMessageVisitor) VisitDeveloper(DeveloperMessage) error { return nil }
func (BaseMessageVisitor) VisitTool(ToolMessage) error           { return nil }
func (BaseMessageVisitor) VisitFunction(FunctionMessage) error   { return nil }
