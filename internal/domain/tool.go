package domain

import "context"

type ToolCategory string

const (
	CategoryRead    ToolCategory = "read"
	CategoryWrite   ToolCategory = "write"
	CategoryCommand ToolCategory = "command"
)

type ToolRegistry interface {
	Get(name string) Tool
	GetAllSchemas() []map[string]any
}

type ToolResult struct {
	Output  string
	IsError bool
	Outcome ToolOutcome
}

type ToolOutcome string

const (
	ToolOutcomeSucceeded         ToolOutcome = "succeeded"
	ToolOutcomeFailed            ToolOutcome = "failed"
	ToolOutcomePermissionDenied  ToolOutcome = "permission_denied"
	ToolOutcomePermissionAborted ToolOutcome = "permission_aborted"
	ToolOutcomeCancelled         ToolOutcome = "cancelled"
)

type Tool interface {
	Name() string
	Description() string
	Category() ToolCategory
	Schema() map[string]any
	Execute(ctx context.Context, args map[string]any) ToolResult
}

type DeferrableTool interface {
	Tool

	ShouldDefer() bool
}

type SystemTool interface {
	Tool

	IsSystemTool() bool
}

func IsSystemTool(t Tool) bool {
	st, ok := t.(SystemTool)
	return ok && st.IsSystemTool()
}
