package svc

import (
	"context"
	"fmt"
	"os"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/prompt"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

const defaultMaxIterations = 30
const DefaultPlanFilePath = ".codefolio/plan.md"

type AgentMode int

const (
	ModeDefault AgentMode = iota
	ModePlan
)

type pendingCall struct {
	id    string
	name  string
	input string
}

type Agent struct {
	MaxIterations int
	Mode          AgentMode
	driver        llm.Driver
	execFactory   domain.ExecutorFactory
	toolRegistry  domain.ToolRegistry
}

func NewAgent(driver llm.Driver, execFactory domain.ExecutorFactory, toolRegistry domain.ToolRegistry) *Agent {
	return &Agent{
		MaxIterations: defaultMaxIterations,
		driver:        driver,
		execFactory:   execFactory,
		toolRegistry:  toolRegistry,
	}
}

func (a *Agent) Run(ctx context.Context, session domain.Session, cfg *conf.Global, cb domain.EventVisitor) error {
	var totalInputTokens int64
	var totalOutputTokens int64
	iter := 0

	maxIter := a.MaxIterations
	if maxIter <= 0 {
		maxIter = defaultMaxIterations
	}

	for {
		iter++
		if iter > maxIter {
			cb.VisitError(domain.ErrorEvent{Err: fmt.Errorf("max agent iterations (%d) exceeded", maxIter)})
			return fmt.Errorf("max agent iterations (%d) exceeded", maxIter)
		}

		if a.Mode == ModePlan {
			planExists := fileExists(DefaultPlanFilePath)
			reminder := prompt.BuildPlanModeReminder(DefaultPlanFilePath, planExists, iter)
			session.AddSystemMessage(reminder)
		}

		messages := ChatMessagesToLLM(session.Messages())
		schemas := a.toolRegistry.GetAllSchemas()

		deltaCh, errCh := a.driver.Stream(ctx, messages,
			llm.WithModel(cfg.Model),
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

				c := &agentCollector{cb: cb, toolCalls: &toolCalls, session: session}
				if err := delta.Accept(c); err != nil {
					cb.VisitError(domain.ErrorEvent{Err: err})
					return err
				}
				totalInputTokens += c.inputTokens
				totalOutputTokens += c.outputTokens
				if c.stopReason == "tool_use" {
					hasToolUse = true
					session.FinishAssistantMessage()
				}

			case err, ok := <-errCh:
				if ok && err != nil {
					cb.VisitError(domain.ErrorEvent{Err: err})
					return err
				}
			}
		}

		select {
		case err, ok := <-errCh:
			if ok && err != nil {
				cb.VisitError(domain.ErrorEvent{Err: err})
				return err
			}
		default:
		}

		cb.VisitUsage(domain.UsageEvent{
			InputTokens:  totalInputTokens,
			OutputTokens: totalOutputTokens,
		})

		if !hasToolUse {
			session.FinishAssistantMessage()
			cb.VisitLoopComplete(domain.LoopCompleteEvent{TotalTurns: iter})
			return nil
		}

		for _, tc := range toolCalls {
			last := session.LastMessage()
			if last == nil {
				break
			}
			for i := range last.ToolCalls {
				if last.ToolCalls[i].ID == tc.id {
					last.ToolCalls[i].Input = tc.input
					break
				}
			}
		}

		exec := a.execFactory(a.toolRegistry)
		for _, tc := range toolCalls {
			exec.Submit(ctx, tc.id, tc.name, tc.input)
		}

		results := exec.CollectResults()
		for _, r := range results {
			session.AddToolResultMessage(r.ID, r.Output, r.IsError)
			cb.VisitToolResult(r)
		}

		cb.VisitTurnComplete(domain.TurnCompleteEvent{Turn: iter})
		session.StartAssistantMessage()
	}
}

type agentCollector struct {
	llm.BaseDeltaVisitor
	cb           domain.EventVisitor
	toolCalls    *[]pendingCall
	session      domain.Session
	stopReason   string
	inputTokens  int64
	outputTokens int64
}

func (c *agentCollector) VisitMessage(d llm.MessageDelta) error {
	if d.Content != "" {
		c.cb.VisitStream(domain.StreamEvent{Content: d.Content})
	}
	return nil
}

func (c *agentCollector) VisitThinking(d llm.ThinkingDelta) error {
	if d.Content != "" {
		c.cb.VisitThink(domain.ThinkEvent{Content: d.Content})
	}
	return nil
}

func (c *agentCollector) VisitThinkingStart(d llm.ThinkingStartDelta) error {
	c.cb.VisitThinkStart(domain.ThinkStartEvent{Signature: d.Signature})
	return nil
}

func (c *agentCollector) VisitUsage(d llm.UsageDelta) error {
	c.inputTokens += int64(d.InputTokens)
	c.outputTokens += int64(d.OutputTokens)
	return nil
}

func (c *agentCollector) VisitToolCallStart(d llm.ToolCallStartDelta) error {
	*c.toolCalls = append(*c.toolCalls, pendingCall{
		id:   d.ID,
		name: d.Name,
	})
	c.session.AddToolCallToAssistant(domain.ToolCall{
		ID:   d.ID,
		Name: d.Name,
	})
	c.cb.VisitToolCall(domain.ToolCallEvent{
		ID:   d.ID,
		Name: d.Name,
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
