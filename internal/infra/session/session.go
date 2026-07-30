package session

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

var _ domain.Session = (*Session)(nil)

type Session struct {
	mu        sync.RWMutex
	messages  []domain.ChatMessage
	createdAt time.Time
	msgSeq    int
	context   domain.ContextMetrics
	logPath   string
}

func New() domain.Session {
	return &Session{
		messages:  make([]domain.ChatMessage, 0),
		createdAt: time.Now(),
	}
}

// Load rebuilds the latest message state from an append-only session JSONL
// file. Repeated records for a message are checkpoints; the last valid record
// wins while its original conversation position is retained.
func Load(path string) (domain.Session, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	value := &Session{messages: make([]domain.ChatMessage, 0), createdAt: time.Now()}
	byID := make(map[string]int)
	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 2*1024*1024)
	for scanner.Scan() {
		var record logRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil || record.Version != 1 || record.Kind != "message" || record.Message.ID == "" {
			continue
		}
		if index, ok := byID[record.Message.ID]; ok {
			value.messages[index] = cloneMessage(record.Message)
		} else {
			byID[record.Message.ID] = len(value.messages)
			value.messages = append(value.messages, cloneMessage(record.Message))
		}
		if record.Timestamp.Before(value.createdAt) {
			value.createdAt = record.Timestamp
		}
		if seq := messageSequence(record.Message.ID); seq > value.msgSeq {
			value.msgSeq = seq
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	value.logPath = path
	return value, nil
}

// ConfigurePersistence enables an append-only JSONL transcript for this
// session. Persistence is best-effort: a read-only workspace must not prevent
// an active coding session from continuing in memory.
func (s *Session) ConfigurePersistence(workDir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.logPath != "" || workDir == "" {
		return
	}
	dir := filepath.Join(workDir, ".codefolio", "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	s.logPath = filepath.Join(dir, time.Now().Format("20060102-150405.000000000")+".jsonl")
	for _, message := range s.messages {
		s.persistLocked("message", message)
	}
}

func (s *Session) AddSystemMessage(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msgSeq++
	msg := domain.ChatMessage{
		ID:        sysMsgID(s.msgSeq),
		Role:      llm.RoleSystem,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.messages = append(s.messages, msg)
	s.persistLocked("message", msg)
}

func (s *Session) AddUserMessage(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msgSeq++
	msg := domain.ChatMessage{
		ID:        userMsgID(s.msgSeq),
		Role:      llm.RoleUser,
		Content:   content,
		Timestamp: time.Now(),
	}
	s.messages = append(s.messages, msg)
	s.persistLocked("message", msg)
}

func (s *Session) StartAssistantMessage() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msgSeq++
	msg := domain.ChatMessage{
		ID:        asstMsgID(s.msgSeq),
		Role:      llm.RoleAssistant,
		Content:   "",
		Timestamp: time.Now(),
		Streaming: true,
	}
	s.messages = append(s.messages, msg)
	s.persistLocked("message", msg)
}

func (s *Session) AppendDelta(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	if last.Streaming {
		last.Content += content
	}
}

func (s *Session) AppendThinkingDelta(content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	if last.Streaming {
		last.Thinking += content
	}
}

func (s *Session) FinishAssistantMessage() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	if !last.Streaming {
		return
	}
	last.Streaming = false
	s.persistLocked("message", *last)
}

func (s *Session) AddToolCallToAssistant(tc domain.ToolCall) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	last.ToolCalls = append(last.ToolCalls, tc)
}

func (s *Session) AddToolResultMessage(toolCallID string, content string, isError bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.persistLocked("message", msg)
}

func (s *Session) SetThinkingSignature(signature string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	last.ThinkingSignature = signature
}

func (s *Session) UpdateToolCallInput(id string, input string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.messages) == 0 {
		return
	}
	last := &s.messages[len(s.messages)-1]
	for i := range last.ToolCalls {
		if last.ToolCalls[i].ID == id {
			last.ToolCalls[i].Input = input
			s.persistLocked("message", *last)
			return
		}
	}
}

func (s *Session) LastMessage() *domain.ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.messages) == 0 {
		return nil
	}
	message := cloneMessage(s.messages[len(s.messages)-1])
	return &message
}

func (s *Session) Messages() []domain.ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	messages := make([]domain.ChatMessage, len(s.messages))
	for i := range s.messages {
		messages[i] = cloneMessage(s.messages[i])
	}
	return messages
}

func (s *Session) MessageCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.messages)
}

func (s *Session) SetContextMetrics(metrics domain.ContextMetrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.context = metrics
}

func (s *Session) ContextUsage() domain.ContextMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.context
}

func (s *Session) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.createdAt)
}

type logRecord struct {
	Version   int                `json:"version"`
	Kind      string             `json:"kind"`
	Timestamp time.Time          `json:"timestamp"`
	Message   domain.ChatMessage `json:"message"`
}

func (s *Session) persistLocked(kind string, message domain.ChatMessage) {
	if s.logPath == "" {
		return
	}
	record, err := json.Marshal(logRecord{Version: 1, Kind: kind, Timestamp: time.Now(), Message: message})
	if err != nil {
		return
	}
	file, err := os.OpenFile(s.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.Write(append(record, '\n'))
}

func cloneMessage(message domain.ChatMessage) domain.ChatMessage {
	message.ToolCalls = append([]domain.ToolCall(nil), message.ToolCalls...)
	return message
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

func messageSequence(id string) int {
	separator := strings.LastIndexByte(id, '-')
	if separator < 0 || separator == len(id)-1 {
		return 0
	}
	sequence, err := strconv.Atoi(id[separator+1:])
	if err != nil {
		return 0
	}
	return sequence
}
