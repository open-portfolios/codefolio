package svc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
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
	MaxIterations  int
	Mode           AgentMode
	WorkDir        string
	driver         llm.Driver
	execFactory    domain.ExecutorFactory
	toolRegistry   domain.ToolRegistry
	promptService  domain.PromptService
	contextManager *ContextManager
	memoryService  *MemoryService
}

func NewAgent(driver llm.Driver, execFactory domain.ExecutorFactory, toolRegistry domain.ToolRegistry, promptService domain.PromptService, contextManager *ContextManager, memoryService *MemoryService) *Agent {
	return &Agent{
		MaxIterations:  defaultMaxIterations,
		Mode:           ModePlan,
		driver:         driver,
		execFactory:    execFactory,
		toolRegistry:   toolRegistry,
		promptService:  promptService,
		contextManager: contextManager,
		memoryService:  memoryService,
	}
}

func (a *Agent) Run(ctx context.Context, session domain.Session, cfg *conf.Struct, cb domain.EventVisitor) error {
	var totalInputTokens int64
	var totalOutputTokens int64
	var totalTokens int64
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

		requestMessages := session.ProviderMessages()
		query := ""
		for i := len(requestMessages) - 1; i >= 0; i-- {
			if requestMessages[i].Role == llm.RoleUser {
				query = requestMessages[i].Content
				break
			}
		}
		requestMessages = append(requestMessages, a.memoryService.Context(ctx, a.WorkDir, query, iter == 1, cfg)...)
		if a.Mode == ModePlan {
			planExists := fileExists(DefaultPlanFilePath)
			reminder := a.promptService.BuildPlanModeReminder(DefaultPlanFilePath, planExists, iter)
			// Runtime reminders guide the next request but are not conversation
			// history. Keeping them out of the ledger prevents repeated plan
			// instructions from consuming the compaction budget.
			requestMessages = append(requestMessages, domain.ChatMessage{Role: llm.RoleSystem, Content: reminder})
		}

		schemas := a.toolRegistry.GetAllSchemas()
		preparation, err := a.contextManager.Prepare(ctx, requestMessages, schemas, cfg)
		if err != nil {
			cb.VisitError(domain.ErrorEvent{Err: err})
			return err
		}
		session.SetContextMetrics(preparation.Metrics)
		for _, event := range preparation.Events {
			_ = cb.VisitContext(event)
		}
		messages := ChatMessagesToLLM(preparation.Messages)

		deltaCh, errCh := a.driver.Stream(ctx, messages,
			llm.WithModel(cfg.Model),
			llm.WithMaxTokens(cfg.MaxOutputTokens()),
			llm.WithTools(schemas),
		)

		var toolCalls []pendingCall
		hasToolUse := false
		collector := &agentCollector{cb: cb, toolCalls: &toolCalls, session: session}

	loop:
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case delta, ok := <-deltaCh:
				if !ok {
					break loop
				}

				if err := delta.Accept(collector); err != nil {
					cb.VisitError(domain.ErrorEvent{Err: err})
					return err
				}
				if collector.stopReason == "tool_use" {
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
		totalInputTokens += collector.inputTokens
		totalOutputTokens += collector.outputTokens
		totalTokens += collector.totalTokens
		if collector.hasUsage {
			metrics := preparation.Metrics
			metrics.ActualInputTokens = collector.inputTokens
			session.SetContextMetrics(metrics)
			_ = cb.VisitContext(domain.ContextEvent{Metrics: metrics, Kind: domain.ContextMeasured})
		}

		cb.VisitUsage(domain.UsageEvent{
			InputTokens:  totalInputTokens,
			OutputTokens: totalOutputTokens,
			TotalTokens:  totalTokens,
		})

		if !hasToolUse {
			session.FinishAssistantMessage()
			cb.VisitLoopComplete(domain.LoopCompleteEvent{TotalTurns: iter})
			go a.memoryService.Extract(a.WorkDir, session.Messages(), cfg)
			return nil
		}

		for _, tc := range toolCalls {
			session.UpdateToolCallInput(tc.id, tc.input)
			cb.VisitToolCall(domain.ToolCallEvent{ID: tc.id, Name: tc.name, Input: tc.input})
		}

		exec := a.execFactory(a.toolRegistry, a.WorkDir)
		executionCtx := domain.WithExecutionProfile(ctx, a.executionProfile(), a.planFilePath())
		for _, tc := range toolCalls {
			exec.Submit(executionCtx, tc.id, tc.name, tc.input)
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

func (a *Agent) executionProfile() domain.ExecutionProfile {
	if a.Mode == ModePlan {
		return domain.ProfilePlan
	}
	return domain.ProfileBuild
}

func (a *Agent) planFilePath() string {
	if a.WorkDir == "" {
		return DefaultPlanFilePath
	}
	return filepath.Join(a.WorkDir, DefaultPlanFilePath)
}

type agentCollector struct {
	llm.BaseDeltaVisitor

	cb           domain.EventVisitor
	toolCalls    *[]pendingCall
	session      domain.Session
	stopReason   string
	inputTokens  int64
	outputTokens int64
	totalTokens  int64
	hasUsage     bool
}

func (c *agentCollector) VisitMessage(d llm.MessageDelta) error {
	if d.Content != "" {
		c.session.AppendDelta(d.Content)
		c.cb.VisitStream(domain.StreamEvent{Content: d.Content})
	}
	return nil
}

func (c *agentCollector) VisitThinking(d llm.ThinkingDelta) error {
	if d.Content != "" {
		c.session.AppendThinkingDelta(d.Content)
		c.cb.VisitThink(domain.ThinkEvent{Content: d.Content})
	}
	return nil
}

func (c *agentCollector) VisitThinkingStart(d llm.ThinkingStartDelta) error {
	c.session.SetThinkingSignature(d.Signature)
	c.cb.VisitThinkStart(domain.ThinkStartEvent{Signature: d.Signature})
	return nil
}

func (c *agentCollector) VisitUsage(d llm.UsageDelta) error {
	if !d.Final {
		return nil
	}
	c.inputTokens = int64(d.InputTokens)
	c.outputTokens = int64(d.OutputTokens)
	c.totalTokens = int64(d.TotalTokens)
	c.hasUsage = true
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
