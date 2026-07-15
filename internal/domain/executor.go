package domain

import "context"

type Executor interface {
	Submit(ctx context.Context, toolID, toolName, input string)
	CollectResults() []ToolResultEvent
}

type ExecutorFactory func(registry ToolRegistry) Executor
