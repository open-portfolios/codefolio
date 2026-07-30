package domain

import "context"

type PermissionEffect string

const (
	PermissionAllow PermissionEffect = "allow"
	PermissionDeny  PermissionEffect = "deny"
	PermissionAsk   PermissionEffect = "ask"
)

// ExecutionProfile selects the prompt and permission policy for a turn.
// Profiles are carried to the executor so policy is enforced at the tool
// boundary rather than relying only on model instructions.
type ExecutionProfile string

const (
	ProfileBuild ExecutionProfile = "build"
	ProfilePlan  ExecutionProfile = "plan"
)

type ToolInvocation struct {
	ID       string
	Name     string
	Category ToolCategory
	Input    string
	Args     map[string]any
	WorkDir  string
	Profile  ExecutionProfile
	PlanFile string
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
type executionProfileKey struct{}

func WithExecutionWorkDir(ctx context.Context, workDir string) context.Context {
	return context.WithValue(ctx, executionWorkDirKey{}, workDir)
}

func ExecutionWorkDir(ctx context.Context) string {
	workDir, _ := ctx.Value(executionWorkDirKey{}).(string)
	return workDir
}

func WithExecutionProfile(ctx context.Context, profile ExecutionProfile, planFile string) context.Context {
	return context.WithValue(ctx, executionProfileKey{}, executionProfile{profile: profile, planFile: planFile})
}

func ExecutionProfileFromContext(ctx context.Context) (ExecutionProfile, string) {
	value, _ := ctx.Value(executionProfileKey{}).(executionProfile)
	return value.profile, value.planFile
}

type executionProfile struct {
	profile  ExecutionProfile
	planFile string
}
