package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/open-portfolios/codefolio/internal/domain"
)

type executor struct {
	registry   domain.ToolRegistry
	authorizer domain.Authorizer
	workDir    string
	mu         sync.Mutex
	pending    []pendingTool
	wg         sync.WaitGroup
}

type pendingTool struct {
	toolID   string
	toolName string
	input    string
	result   *toolExecResult
	done     bool
}

type toolExecResult struct {
	toolID   string
	toolName string
	output   string
	isError  bool
	outcome  domain.ToolOutcome
	elapsed  time.Duration
}

func NewExecutor(registry domain.ToolRegistry, authorizer domain.Authorizer, workDir string) domain.Executor {
	return &executor{registry: registry, authorizer: authorizer, workDir: workDir}
}

func (e *executor) Submit(ctx context.Context, toolID, toolName, input string) {
	e.mu.Lock()
	idx := len(e.pending)
	e.pending = append(e.pending, pendingTool{
		toolID:   toolID,
		toolName: toolName,
		input:    input,
	})
	e.mu.Unlock()

	e.wg.Go(func() {
		result := e.execute(ctx, toolID, toolName, input)
		e.mu.Lock()
		e.pending[idx].result = &result
		e.pending[idx].done = true
		e.mu.Unlock()
	})
}

func (e *executor) CollectResults() []domain.ToolResultEvent {
	e.wg.Wait()

	e.mu.Lock()
	defer e.mu.Unlock()

	results := make([]domain.ToolResultEvent, len(e.pending))
	for i, p := range e.pending {
		if p.result != nil {
			results[i] = domain.ToolResultEvent{
				ID:      p.result.toolID,
				Name:    p.result.toolName,
				Output:  p.result.output,
				IsError: p.result.isError,
				Outcome: p.result.outcome,
				Elapsed: p.result.elapsed,
			}
		}
	}
	return results
}

func (e *executor) execute(ctx context.Context, toolID, toolName, input string) toolExecResult {
	start := time.Now()

	t := e.registry.Get(toolName)
	if t == nil {
		return toolExecResult{
			toolID:   toolID,
			toolName: toolName,
			output:   fmt.Sprintf("Error: unknown tool %q", toolName),
			isError:  true,
			elapsed:  time.Since(start),
		}
	}

	var args map[string]any
	if input != "" {
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return toolExecResult{
				toolID:   toolID,
				toolName: toolName,
				output:   fmt.Sprintf("Error: invalid tool arguments: %s", err),
				isError:  true,
				elapsed:  time.Since(start),
			}
		}
	}
	normalizePathArgs(args, e.workDir)
	if err := ctx.Err(); err != nil {
		return toolExecResult{toolID: toolID, toolName: toolName, output: "Error: tool execution cancelled", isError: true, outcome: domain.ToolOutcomeCancelled, elapsed: time.Since(start)}
	}
	if e.authorizer != nil {
		decision := e.authorizer.Authorize(ctx, domain.ToolInvocation{ID: toolID, Name: toolName, Category: t.Category(), Input: input, Args: args, WorkDir: e.workDir})
		if decision.Effect != domain.PermissionAllow {
			outcome := domain.ToolOutcomePermissionDenied
			if decision.Reason == "approval cancelled" {
				outcome = domain.ToolOutcomePermissionAborted
			}
			return toolExecResult{toolID: toolID, toolName: toolName, output: "Permission denied: " + decision.Reason, isError: true, outcome: outcome, elapsed: time.Since(start)}
		}
	}
	if err := ctx.Err(); err != nil {
		return toolExecResult{toolID: toolID, toolName: toolName, output: "Error: tool execution cancelled", isError: true, outcome: domain.ToolOutcomeCancelled, elapsed: time.Since(start)}
	}

	result := t.Execute(domain.WithExecutionWorkDir(ctx, e.workDir), args)
	outcome := result.Outcome
	if outcome == "" {
		outcome = domain.ToolOutcomeSucceeded
		if result.IsError {
			outcome = domain.ToolOutcomeFailed
		}
	}
	return toolExecResult{
		toolID:   toolID,
		toolName: toolName,
		output:   result.Output,
		isError:  result.IsError,
		outcome:  outcome,
		elapsed:  time.Since(start),
	}
}

func normalizePathArgs(args map[string]any, workDir string) {
	if workDir == "" {
		return
	}
	for _, key := range []string{"file_path", "path"} {
		value, ok := args[key].(string)
		if !ok || value == "" || filepath.IsAbs(value) {
			continue
		}
		args[key] = filepath.Join(workDir, value)
	}
}
