package domain

import "time"

type ToolCall struct {
	ID    string
	Name  string
	Input string
}

type ChatMessage struct {
	ID                string
	Role              string
	Content           string
	Thinking          string
	ThinkingSignature string
	ThinkingExpanded  bool
	Timestamp         time.Time
	Streaming         bool
	Error             string
	IsError           bool
	ToolCalls         []ToolCall
	ToolCallID        string
}
