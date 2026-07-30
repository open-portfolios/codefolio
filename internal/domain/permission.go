package domain

import "context"

type PermissionEffect string

const (
	PermissionAllow PermissionEffect = "allow"
	PermissionDeny  PermissionEffect = "deny"
	PermissionAsk   PermissionEffect = "ask"
)

type ToolInvocation struct {
	ID       string
	Name     string
	Category ToolCategory
	Input    string
	Args     map[string]any
	WorkDir  string
}

type PermissionDecision struct {
	Effect PermissionEffect
	Reason string
}

// Authorizer decides whether a tool invocation may run before it reaches the
// tool implementation.
type Authorizer interface {
	Authorize(context.Context, ToolInvocation) PermissionDecision
}

type executionWorkDirKey struct{}

func WithExecutionWorkDir(ctx context.Context, workDir string) context.Context {
	return context.WithValue(ctx, executionWorkDirKey{}, workDir)
}

func ExecutionWorkDir(ctx context.Context) string {
	workDir, _ := ctx.Value(executionWorkDirKey{}).(string)
	return workDir
}
