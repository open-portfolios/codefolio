package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/open-portfolios/codefolio/internal/domain"
)

type executor struct {
	registry domain.ToolRegistry
	mu       sync.Mutex
	pending  []pendingTool
	wg       sync.WaitGroup
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
	elapsed  time.Duration
}

func NewExecutor(registry domain.ToolRegistry) domain.Executor {
	return &executor{registry: registry}
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

	result := t.Execute(ctx, args)
	return toolExecResult{
		toolID:   toolID,
		toolName: toolName,
		output:   result.Output,
		isError:  result.IsError,
		elapsed:  time.Since(start),
	}
}
