package domain

import "time"

type Session interface {
	AddSystemMessage(content string)
	AddUserMessage(content string)
	StartAssistantMessage()
	AppendDelta(content string)
	AppendThinkingDelta(content string)
	FinishAssistantMessage()
	AddToolCallToAssistant(tc ToolCall)
	AddToolResultMessage(toolCallID string, content string, isError bool)
	SetThinkingSignature(signature string)
	LastMessage() *ChatMessage
	Messages() []ChatMessage
	MessageCount() int
	SetContextMetrics(ContextMetrics)
	ContextUsage() ContextMetrics
	Duration() time.Duration
}
