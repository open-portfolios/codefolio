package llm

import "errors"

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleDeveloper = "developer"
	RoleSystem    = "system"
	RoleTool      = "tool"
	RoleFunction  = "function"
)

var (
	ErrMalformedToolMessage     = errors.New("malformed tool message")
	ErrMalformedFunctionMessage = errors.New("malformed function message")
)

type Message interface {
	Role() string
	Content() string
}

type ToolCallMessage interface {
	Message

	ToolCallID() string
}

type FunctionMessage interface {
	Message

	Name() string
}

type ToolCallInfo struct {
	ID    string
	Name  string
	Input string
}

type MessageWithToolCalls interface {
	Message

	ToolCalls() []ToolCallInfo
}

type MessageWithThinking interface {
	Message

	Thinking() string
	ThinkingSignature() string
}
