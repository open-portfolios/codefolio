package session

import (
	"time"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

var _ domain.Session = (*Session)(nil)

type Session struct {
	messages  []domain.ChatMessage
	createdAt time.Time
	msgSeq    int
	context   domain.ContextMetrics
}

func New() domain.Session {
	return &Session{
		messages:  make([]domain.ChatMessage, 0),
		createdAt: time.Now(),
	}
}

func (s *Session) AddSystemMessage(content string) {
	s.msgSeq++
	msg := domain.ChatMessage{
		ID:        sysMsgID(s.msgSeq),
		Role:      llm.RoleSystem,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.messages = append(s.messages, msg)
}

func (s *Session) AddUserMessage(content string) {
	s.msgSeq++
	msg := domain.ChatMessage{
		ID:        userMsgID(s.msgSeq),
		Role:      llm.RoleUser,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.messages = append(s.messages, msg)
}

func (s *Session) StartAssistantMessage() {
	s.msgSeq++
	msg := domain.ChatMessage{
		ID:        asstMsgID(s.msgSeq),
		Role:      llm.RoleAssistant,
		Content:   "",
		Timestamp: time.Now(),
		Streaming: true,
	}
	s.messages = append(s.messages, msg)
}

func (s *Session) AppendDelta(content string) {
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	if last.Streaming {
		last.Content += content
	}
}

func (s *Session) AppendThinkingDelta(content string) {
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	if last.Streaming {
		last.Thinking += content
	}
}

func (s *Session) FinishAssistantMessage() {
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	last.Streaming = false
}

func (s *Session) AddToolCallToAssistant(tc domain.ToolCall) {
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	last.ToolCalls = append(last.ToolCalls, tc)
}

func (s *Session) AddToolResultMessage(toolCallID string, content string, isError bool) {
	s.msgSeq++
	msg := domain.ChatMessage{
		ID:         toolMsgID(s.msgSeq),
		Role:       llm.RoleTool,
		Content:    content,
		IsError:    isError,
		ToolCallID: toolCallID,
		Timestamp:  time.Now(),
	}
	s.messages = append(s.messages, msg)
}

func (s *Session) SetThinkingSignature(signature string) {
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	last.ThinkingSignature = signature
}

func (s *Session) LastMessage() *domain.ChatMessage {
	if len(s.messages) == 0 {
		return nil
	}
	return &s.messages[len(s.messages)-1]
}

func (s *Session) Messages() []domain.ChatMessage {
	return s.messages
}

func (s *Session) MessageCount() int {
	return len(s.messages)
}

func (s *Session) SetContextMetrics(metrics domain.ContextMetrics) { s.context = metrics }

func (s *Session) ContextUsage() domain.ContextMetrics { return s.context }

func (s *Session) Duration() time.Duration {
	return time.Since(s.createdAt)
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
