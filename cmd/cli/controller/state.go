package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cylixlee/tux/app"
	"github.com/cylixlee/tux/builtin"
	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/approval"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
	"github.com/open-portfolios/codefolio/internal/svc"
)

type StreamState uint8

const (
	StreamIdle StreamState = iota
	StreamRunning
)

type Tool struct {
	ID       string
	Name     string
	Input    string
	Output   string
	Done     bool
	IsError  bool
	Elapsed  time.Duration
	Outcome  domain.ToolOutcome
	Expanded bool
}

type Message struct {
	ID               string
	Role             string
	Content          string
	Thinking         string
	ThinkingExpanded bool
	Streaming        bool
	Error            string
	Tools            []*Tool
}

type AskState struct {
	Request    *askuser.Request
	Question   int
	Selected   int
	Selections []int
}

type ApprovalState struct {
	Request *approval.Request
	Queue   []*approval.Request
}

type Controller struct {
	cfg        *conf.Global
	agent      *svc.Agent
	session    domain.Session
	askUserCh  <-chan askuser.Request
	approvalCh <-chan *approval.Request

	mu              sync.RWMutex
	runtime         *app.App
	invalidate      func()
	setAskOpen      func(bool)
	setApprovalOpen func(bool)
	messages        []*Message
	streaming       StreamState
	cancelling      bool
	runID           uint64
	cancel          context.CancelFunc
	history         []string
	historyAt       int
	draft           string
	inputTokens     int64
	outputTokens    int64
	ask             AskState
	approval        ApprovalState
}

func New(cfg *conf.Global, agent *svc.Agent, session domain.Session, askUserCh chan askuser.Request, approvalCh chan *approval.Request) *Controller {
	return &Controller{cfg: cfg, agent: agent, session: session, askUserCh: askUserCh, approvalCh: approvalCh, historyAt: -1}
}

func (c *Controller) Attach(runtime *app.App, invalidate func(), setAskOpen func(bool), setApprovalOpen func(bool)) {
	c.mu.Lock()
	c.runtime, c.invalidate, c.setAskOpen, c.setApprovalOpen = runtime, invalidate, setAskOpen, setApprovalOpen
	c.mu.Unlock()
	go c.askLoop()
	go c.approvalLoop()
}

func (c *Controller) ModelName() string { return c.cfg.Model }

func (c *Controller) Messages() []*Message { return c.messages }

func (c *Controller) Status() string {
	if c.cancelling {
		return "cancelling"
	}
	if c.streaming == StreamRunning {
		return "working"
	}
	return "idle"
}

func (c *Controller) InputTokens() int64      { return c.inputTokens }
func (c *Controller) OutputTokens() int64     { return c.outputTokens }
func (c *Controller) Running() bool           { return c.streaming == StreamRunning }
func (c *Controller) Cancelling() bool        { return c.cancelling }
func (c *Controller) Ask() AskState           { return c.ask }
func (c *Controller) Approval() ApprovalState { return c.approval }

func (c *Controller) Tick() {
	if c.streaming == StreamRunning {
		c.invalidate()
	}
}

func (c *Controller) Start(content string, editor *builtin.TextareaState) {
	c.history = append(c.history, content)
	c.historyAt, c.draft = len(c.history), ""
	c.session.AddUserMessage(content)
	c.session.StartAssistantMessage()
	c.messages = append(c.messages, &Message{ID: fmt.Sprintf("user-%d", len(c.messages)+1), Role: "user", Content: content})
	c.messages = append(c.messages, &Message{ID: fmt.Sprintf("assistant-%d", len(c.messages)+1), Role: "assistant", Streaming: true})
	c.streaming, c.cancelling = StreamRunning, false
	c.runID++
	runID := c.runID
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	*editor = builtin.TextareaState{PreferredColumn: -1, CursorOn: true}
	runtime := c.runtime
	go func() {
		visitor := &visitor{controller: c, runtime: runtime, runID: runID, ctx: ctx}
		defer func() {
			c.session.FinishAssistantMessage()
			if runtime != nil {
				_ = runtime.Dispatch(context.Background(), func() {
					if runID == c.runID {
						c.cancelling = false
						c.invalidate()
					}
				})
			}
		}()
		if err := c.agent.Run(ctx, c.session, c.cfg, visitor); err != nil {
			_ = visitor.post(func() { c.applyError(runID, err) })
		}
	}()
	c.invalidate()
}

func (c *Controller) Cancel() {
	if c.streaming != StreamRunning || c.cancel == nil {
		return
	}
	c.cancel()
	c.cancel = nil
	c.finish("interrupted")
	c.cancelling = true
	c.denyAllApprovals()
	c.invalidate()
}

func (c *Controller) Shutdown() {
	c.Cancel()
	c.denyAllApprovals()
}

func (c *Controller) HistoryUp(value string) string {
	if len(c.history) == 0 || c.historyAt <= 0 {
		return value
	}
	if c.historyAt == len(c.history) {
		c.draft = value
	}
	c.historyAt--
	return c.history[c.historyAt]
}

func (c *Controller) HistoryDown() string {
	if c.historyAt >= len(c.history) {
		return ""
	}
	c.historyAt++
	if c.historyAt == len(c.history) {
		return c.draft
	}
	return c.history[c.historyAt]
}

func (c *Controller) ToggleLastThinking() {
	for i := len(c.messages) - 1; i >= 0; i-- {
		if c.messages[i].Thinking != "" {
			c.messages[i].ThinkingExpanded = !c.messages[i].ThinkingExpanded
			c.invalidate()
			return
		}
	}
}

func (c *Controller) ToggleThinking(id string) {
	for _, m := range c.messages {
		if m.ID == id {
			m.ThinkingExpanded = !m.ThinkingExpanded
			c.invalidate()
			return
		}
	}
}

func (c *Controller) ToggleTool(messageID, toolID string) {
	for _, m := range c.messages {
		if m.ID == messageID {
			for _, tool := range m.Tools {
				if tool.ID == toolID {
					tool.Expanded = !tool.Expanded
					c.invalidate()
					return
				}
			}
		}
	}
}

// StartAssistantSegment separates the next model iteration from the assistant
// activity that invoked tools. This is presentation-only; the domain session
// continues to own the provider's assistant-message boundaries.
func (c *Controller) StartAssistantSegment() {
	c.messages = append(c.messages, &Message{ID: fmt.Sprintf("assistant-%d", len(c.messages)+1), Role: "assistant", Streaming: true})
	c.invalidate()
}

func (c *Controller) currentAssistant() *Message {
	for i := len(c.messages) - 1; i >= 0; i-- {
		if c.messages[i].Role == "assistant" {
			return c.messages[i]
		}
	}
	return nil
}

func (c *Controller) finish(reason string) {
	if c.streaming != StreamRunning {
		return
	}
	if message := c.currentAssistant(); message != nil {
		message.Streaming = false
		message.Error = reason
	}
	c.streaming, c.cancel = StreamIdle, nil
}

func (c *Controller) applyError(runID uint64, err error) {
	if runID == c.runID && c.streaming == StreamRunning {
		c.finish(err.Error())
		c.invalidate()
	}
}
