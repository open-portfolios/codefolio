package domain

import (
	"time"

	"github.com/open-portfolios/codefolio/pkg/llm"
)

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

type Session struct {
	Messages  []ChatMessage
	CreatedAt time.Time
	msgSeq    int
}

func NewSession() *Session {
	return &Session{
		Messages:  make([]ChatMessage, 0),
		CreatedAt: time.Now(),
	}
}

func (s *Session) AddSystemMessage(content string) {
	s.msgSeq++
	msg := ChatMessage{
		ID:        sysMsgID(s.msgSeq),
		Role:      llm.RoleSystem,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.Messages = append(s.Messages, msg)
}

func (s *Session) AddUserMessage(content string) {
	s.msgSeq++
	msg := ChatMessage{
		ID:        userMsgID(s.msgSeq),
		Role:      llm.RoleUser,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.Messages = append(s.Messages, msg)
}

func (s *Session) StartAssistantMessage() {
	s.msgSeq++
	msg := ChatMessage{
		ID:        asstMsgID(s.msgSeq),
		Role:      llm.RoleAssistant,
		Content:   "",
		Timestamp: time.Now(),
		Streaming: true,
	}
	s.Messages = append(s.Messages, msg)
}

func (s *Session) AppendDelta(content string) {
	if len(s.Messages) == 0 {
		return
	}
	last := &s.Messages[len(s.Messages)-1]
	if last.Streaming {
		last.Content += content
	}
}

func (s *Session) AppendThinkingDelta(content string) {
	if len(s.Messages) == 0 {
		return
	}
	last := &s.Messages[len(s.Messages)-1]
	if last.Streaming {
		last.Thinking += content
	}
}

func (s *Session) FinishAssistantMessage() {
	if len(s.Messages) == 0 {
		return
	}
	last := &s.Messages[len(s.Messages)-1]
	last.Streaming = false
}

func (s *Session) AddToolCallToAssistant(toolCall ToolCall) {
	if len(s.Messages) == 0 {
		return
	}
	last := &s.Messages[len(s.Messages)-1]
	last.ToolCalls = append(last.ToolCalls, toolCall)
}

func (s *Session) AddToolResultMessage(toolCallID string, content string, isError bool) {
	s.msgSeq++
	msg := ChatMessage{
		ID:         toolMsgID(s.msgSeq),
		Role:       llm.RoleTool,
		Content:    content,
		IsError:    isError,
		ToolCallID: toolCallID,
		Timestamp:  time.Now(),
	}
	s.Messages = append(s.Messages, msg)
}

func (s *Session) SetThinkingSignature(signature string) {
	if len(s.Messages) == 0 {
		return
	}
	last := &s.Messages[len(s.Messages)-1]
	last.ThinkingSignature = signature
}

func (s *Session) LastMessage() *ChatMessage {
	if len(s.Messages) == 0 {
		return nil
	}
	return &s.Messages[len(s.Messages)-1]
}

func (s *Session) MessageCount() int {
	return len(s.Messages)
}

func (s *Session) ContextUsage() float64 {
	return 0.0
}

func (s *Session) Duration() time.Duration {
	return time.Since(s.CreatedAt)
}

func userMsgID(seq int) string { return "u-" + itoa(seq) }
func asstMsgID(seq int) string { return "a-" + itoa(seq) }
func sysMsgID(seq int) string  { return "s-" + itoa(seq) }
func toolMsgID(seq int) string { return "t-" + itoa(seq) }

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
