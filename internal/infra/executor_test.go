package infra

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/open-portfolios/codefolio/internal/domain"
)

type testTool struct{ called bool }

func (t *testTool) Name() string                  { return "WriteFile" }
func (t *testTool) Description() string           { return "" }
func (t *testTool) Category() domain.ToolCategory { return domain.CategoryWrite }
func (t *testTool) Schema() map[string]any        { return nil }
func (t *testTool) Execute(context.Context, map[string]any) domain.ToolResult {
	t.called = true
	return domain.ToolResult{Output: "written"}
}

type testRegistry struct{ tool domain.Tool }

func (r testRegistry) Get(string) domain.Tool          { return r.tool }
func (r testRegistry) GetAllSchemas() []map[string]any { return nil }

type denyAuthorizer struct{}

func (denyAuthorizer) Authorize(context.Context, domain.ToolInvocation) domain.PermissionDecision {
	return domain.PermissionDecision{Effect: domain.PermissionDeny, Reason: "policy denied"}
}

func TestExecutorDoesNotExecuteDeniedTool(t *testing.T) {
	tool := &testTool{}
	exec := NewExecutor(testRegistry{tool: tool}, denyAuthorizer{}, t.TempDir())
	exec.Submit(context.Background(), "tool-1", tool.Name(), `{"file_path":"notes.txt","content":"test"}`)
	results := exec.CollectResults()
	if tool.called {
		t.Fatal("denied tool executed")
	}
	if len(results) != 1 || !results[0].IsError || results[0].Outcome != domain.ToolOutcomePermissionDenied {
		t.Fatalf("unexpected result: %#v", results)
	}
}

func TestExecutorNormalizesRelativeFilePaths(t *testing.T) {
	tool := &testTool{}
	authorizer := &recordingAuthorizer{}
	workDir := t.TempDir()
	exec := NewExecutor(testRegistry{tool: tool}, authorizer, workDir)
	exec.Submit(context.Background(), "tool-1", tool.Name(), `{"file_path":"notes.txt","content":"test"}`)
	exec.CollectResults()
	if authorizer.invocation.Args["file_path"] != filepath.Join(workDir, "notes.txt") {
		t.Fatalf("path was not normalized: %#v", authorizer.invocation.Args)
	}
}

type recordingAuthorizer struct{ invocation domain.ToolInvocation }

func (a *recordingAuthorizer) Authorize(_ context.Context, invocation domain.ToolInvocation) domain.PermissionDecision {
	a.invocation = invocation
	return domain.PermissionDecision{Effect: domain.PermissionAllow}
}
