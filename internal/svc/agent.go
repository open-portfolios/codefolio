package svc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

const maxAgentIterations = 30

type pendingCall struct {
	id    string
	name  string
	input string
}

type EventType string

const (
	EventDelta          EventType = "delta"
	EventThinking       EventType = "thinking"
	EventThinkingStart  EventType = "thinking_start"
	EventToolStart      EventType = "tool_start"
	EventToolDone    EventType = "tool_done"
	EventDone        EventType = "done"
	EventError       EventType = "error"
)

type Event struct {
	Type      EventType
	Content   string
	ToolName  string
	ToolInput string
	ToolID    string
	Err       error
}

type Callback func(Event)

type Agent struct {
	Protocol string
}

func NewAgent(protocol string) *Agent {
	return &Agent{Protocol: protocol}
}

func (a *Agent) Run(ctx context.Context, driver llm.Driver, session *domain.Session, registry *tools.Registry, model string, cb Callback) error {
	iter := 0
	for {
		iter++
		if iter > maxAgentIterations {
			cb(Event{Type: EventError, Err: fmt.Errorf("max agent iterations (%d) exceeded", maxAgentIterations)})
			return fmt.Errorf("max agent iterations (%d) exceeded", maxAgentIterations)
		}

		messages := session.ToLLMMessages()
		schemas := registry.GetAllSchemas(a.Protocol)

		deltaCh, errCh := driver.Stream(ctx, messages,
			llm.WithModel(model),
			llm.WithTools(schemas),
		)

		var toolCalls []pendingCall
		hasToolUse := false

	loop:
		for {
			select {
			case delta, ok := <-deltaCh:
				if !ok {
					break loop
				}

				c := &agentCollector{cb: cb, toolCalls: &toolCalls}
				if err := delta.Accept(c); err != nil {
					cb(Event{Type: EventError, Err: err})
					return err
				}
				if c.stopReason == "tool_use" || c.stopReason == "end_turn" {
					if c.stopReason == "tool_use" {
						hasToolUse = true
						session.FinishAssistantMessage()
					}
				}

			case err, ok := <-errCh:
				if ok && err != nil {
					cb(Event{Type: EventError, Err: err})
					return err
				}
			}
		}

		select {
		case err, ok := <-errCh:
			if ok && err != nil {
				cb(Event{Type: EventError, Err: err})
				return err
			}
		default:
		}

		if !hasToolUse {
			session.FinishAssistantMessage()
			cb(Event{Type: EventDone})
			return nil
		}

		for _, tc := range toolCalls {
			session.AddToolCallToAssistant(domain.ToolCall{
				ID:    tc.id,
				Name:  tc.name,
				Input: tc.input,
			})
		}

		for _, tc := range toolCalls {
			t := registry.Get(tc.name)
			if t == nil {
				errMsg := fmt.Sprintf("Error: unknown tool %q", tc.name)
				session.AddToolResultMessage(tc.id, errMsg)
				cb(Event{Type: EventToolDone, ToolID: tc.id, Content: errMsg})
				continue
			}

			cb(Event{Type: EventToolStart, ToolName: tc.name, ToolInput: tc.input, ToolID: tc.id})

			var args map[string]any
			if tc.input != "" {
				if err := json.Unmarshal([]byte(tc.input), &args); err != nil {
					errMsg := fmt.Sprintf("Error: invalid tool arguments: %s", err)
					session.AddToolResultMessage(tc.id, errMsg)
					cb(Event{Type: EventToolDone, ToolID: tc.id, Content: errMsg})
					continue
				}
			}

			result := t.Execute(ctx, args)
			session.AddToolResultMessage(tc.id, result.Output)
			cb(Event{Type: EventToolDone, ToolID: tc.id, Content: result.Output, ToolName: tc.name})
		}
		session.StartAssistantMessage()
	}
}

type agentCollector struct {
	llm.BaseDeltaVisitor
	cb         Callback
	toolCalls  *[]pendingCall
	stopReason string
}

func (c *agentCollector) VisitMessage(d llm.MessageDelta) error {
	if d.Content != "" {
		c.cb(Event{Type: EventDelta, Content: d.Content})
	}
	return nil
}

func (c *agentCollector) VisitThinking(d llm.ThinkingDelta) error {
	if d.Content != "" {
		c.cb(Event{Type: EventThinking, Content: d.Content})
	}
	return nil
}

func (c *agentCollector) VisitThinkingStart(d llm.ThinkingStartDelta) error {
	c.cb(Event{Type: EventThinkingStart, Content: d.Signature})
	return nil
}

func (c *agentCollector) VisitToolCallStart(d llm.ToolCallStartDelta) error {
	*c.toolCalls = append(*c.toolCalls, pendingCall{
		id:   d.ID,
		name: d.Name,
	})
	return nil
}

func (c *agentCollector) VisitToolCallInput(d llm.ToolCallInputDelta) error {
	if len(*c.toolCalls) > d.Index {
		(*c.toolCalls)[d.Index].input += d.Input
	}
	return nil
}

func (c *agentCollector) VisitStreamStop(d llm.StreamStopDelta) error {
	c.stopReason = d.StopReason
	return nil
}
