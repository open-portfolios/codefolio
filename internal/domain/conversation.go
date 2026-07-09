package domain

import (
	"time"

	"github.com/open-portfolios/codefolio/pkg/llm"
)

type ChatMessage struct {
	ID               string
	Role             string
	Content          string
	Thinking         string
	ThinkingExpanded bool
	Timestamp        time.Time
	Streaming        bool
}

func (m ChatMessage) ToLLMMessage() llm.Message {
	return chatMsgAdapter{m}
}

type chatMsgAdapter struct{ msg ChatMessage }

func (a chatMsgAdapter) Role() string    { return a.msg.Role }
func (a chatMsgAdapter) Content() string { return a.msg.Content }

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
		Role:      "system",
		Content:   content,
		Timestamp: time.Now(),
	}
	s.Messages = append(s.Messages, msg)
}

func (s *Session) AddUserMessage(content string) {
	s.msgSeq++
	msg := ChatMessage{
		ID:        userMsgID(s.msgSeq),
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}
	s.Messages = append(s.Messages, msg)
}

func (s *Session) StartAssistantMessage() {
	s.msgSeq++
	msg := ChatMessage{
		ID:        asstMsgID(s.msgSeq),
		Role:      "assistant",
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

func (s *Session) ToLLMMessages() []llm.Message {
	msgs := make([]llm.Message, 0, len(s.Messages))
	for _, m := range s.Messages {
		if m.Streaming {
			continue
		}
		msgs = append(msgs, m.ToLLMMessage())
	}
	return msgs
}

func userMsgID(seq int) string   { return "u-" + itoa(seq) }
func asstMsgID(seq int) string   { return "a-" + itoa(seq) }
func sysMsgID(seq int) string    { return "s-" + itoa(seq) }

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
