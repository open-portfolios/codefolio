package domain

import "time"

type Session interface {
	ID() string
	ConfigurePersistence(workDir string)
	AddSystemMessage(content string)
	AddUserMessage(content string)
	StartAssistantMessage()
	AppendDelta(content string)
	AppendThinkingDelta(content string)
	FinishAssistantMessage()
	AddToolCallToAssistant(tc ToolCall)
	AddToolResultMessage(toolCallID string, content string, isError bool)
	SetThinkingSignature(signature string)
	UpdateToolCallInput(id string, input string)
	LastMessage() *ChatMessage
	Messages() []ChatMessage
	ProviderMessages() []ChatMessage
	ReplaceProviderMessages(messages []ChatMessage)
	MessageCount() int
	SetContextMetrics(ContextMetrics)
	ContextUsage() ContextMetrics
	Duration() time.Duration
}
