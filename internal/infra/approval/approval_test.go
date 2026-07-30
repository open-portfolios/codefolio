package approval

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/open-portfolios/codefolio/internal/domain"
)

func TestBrokerAllowsWorkspaceReads(t *testing.T) {
	requests := make(chan *Request, 1)
	broker := NewBroker(requests)
	decision := broker.Authorize(context.Background(), invocation(domain.CategoryRead, "ReadFile", map[string]any{"file_path": "README.md"}))
	if decision.Effect != domain.PermissionAllow {
		t.Fatalf("expected read to be allowed, got %#v", decision)
	}
	select {
	case request := <-requests:
		t.Fatalf("unexpected approval request: %#v", request)
	default:
	}
}

func TestBrokerDeniesDangerousCommands(t *testing.T) {
	broker := NewBroker(make(chan *Request, 1))
	decision := broker.Authorize(context.Background(), invocation(domain.CategoryCommand, "Bash", map[string]any{"command": "git reset --hard"}))
	if decision.Effect != domain.PermissionDeny || decision.Reason != "dangerous command blocked" {
		t.Fatalf("expected dangerous command denial, got %#v", decision)
	}
}

func TestBrokerBuildAllowsFileModifyingCommands(t *testing.T) {
	broker := NewBroker(make(chan *Request, 1))
	decision := broker.Authorize(context.Background(), invocation(domain.CategoryCommand, "Bash", map[string]any{"command": "gofmt -w main.go"}))
	if decision.Effect != domain.PermissionAllow || decision.Reason != "file-modifying command allowed by build profile" {
		t.Fatalf("expected build command to be allowed, got %#v", decision)
	}
}

func TestBrokerBuildAllowsWorkspaceWrites(t *testing.T) {
	requests := make(chan *Request, 1)
	broker := NewBroker(requests)
	invocation := invocation(domain.CategoryWrite, "WriteFile", map[string]any{"file_path": "notes.txt"})
	if decision := broker.Authorize(context.Background(), invocation); decision.Effect != domain.PermissionAllow || decision.Reason != "allowed by build profile" {
		t.Fatalf("expected build write to be allowed, got %#v", decision)
	}
	select {
	case request := <-requests:
		t.Fatalf("unexpected approval request: %#v", request)
	default:
	}
}

func TestBrokerDoesNotBlockAfterCancellation(t *testing.T) {
	broker := NewBroker(make(chan *Request))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	decision := broker.Authorize(ctx, invocation(domain.CategoryWrite, "WriteFile", map[string]any{"file_path": "notes.txt"}))
	if decision.Effect != domain.PermissionDeny {
		t.Fatalf("expected cancellation denial, got %#v", decision)
	}
}

func TestBrokerAsksBeforeWritingSensitiveConfig(t *testing.T) {
	requests := make(chan *Request, 1)
	broker := NewBroker(requests)
	result := make(chan domain.PermissionDecision, 1)
	go func() {
		result <- broker.Authorize(context.Background(), invocation(domain.CategoryWrite, "WriteFile", map[string]any{"file_path": filepath.Join(".codefolio", "config.json")}))
	}()
	request := <-requests
	if request.Detail != "sensitive Codefolio configuration requires approval" {
		t.Fatalf("unexpected approval detail: %q", request.Detail)
	}
	request.Resolve(Deny)
	if decision := <-result; decision.Reason != "user denied approval" {
		t.Fatalf("expected user denial, got %#v", decision)
	}
}

func TestBrokerAsksForExternalWorkspaceAccess(t *testing.T) {
	requests := make(chan *Request, 1)
	broker := NewBroker(requests)
	result := make(chan domain.PermissionDecision, 1)
	go func() {
		result <- broker.Authorize(context.Background(), invocation(domain.CategoryWrite, "WriteFile", map[string]any{"file_path": filepath.Join("..", "outside.txt")}))
	}()
	request := <-requests
	if request.Detail != "external workspace access requires approval" {
		t.Fatalf("unexpected approval detail: %q", request.Detail)
	}
	request.Resolve(Deny)
	if decision := <-result; decision.Reason != "user denied approval" {
		t.Fatalf("expected user denial, got %#v", decision)
	}
}

func TestBrokerBuildAllowsBashAndMCPTools(t *testing.T) {
	broker := NewBroker(make(chan *Request, 1))
	for _, input := range []domain.ToolInvocation{
		invocation(domain.CategoryCommand, "Bash", map[string]any{"command": "prettier --write src/app.ts"}),
		invocation(domain.CategoryCommand, "mcp__github__create_issue", map[string]any{}),
	} {
		if decision := broker.Authorize(context.Background(), input); decision.Effect != domain.PermissionAllow {
			t.Fatalf("expected build invocation to be allowed, got %#v", decision)
		}
	}
}

func TestBrokerBuildAsksForNonModifyingBash(t *testing.T) {
	requests := make(chan *Request, 1)
	broker := NewBroker(requests)
	result := make(chan domain.PermissionDecision, 1)
	go func() {
		result <- broker.Authorize(context.Background(), invocation(domain.CategoryCommand, "Bash", map[string]any{"command": "go test ./..."}))
	}()
	request := <-requests
	request.Resolve(Deny)
	if decision := <-result; decision.Reason != "user denied approval" {
		t.Fatalf("expected user denial, got %#v", decision)
	}
}

func TestBrokerAsksForBashReferencingExternalPath(t *testing.T) {
	requests := make(chan *Request, 1)
	broker := NewBroker(requests)
	input := invocation(domain.CategoryCommand, "Bash", map[string]any{"command": "cat " + filepath.Join(t.TempDir(), "outside.txt")})
	result := make(chan domain.PermissionDecision, 1)
	go func() { result <- broker.Authorize(context.Background(), input) }()
	request := <-requests
	if request.Detail != "external workspace access requires approval" {
		t.Fatalf("unexpected approval detail: %q", request.Detail)
	}
	request.Resolve(Deny)
	if decision := <-result; decision.Reason != "user denied approval" {
		t.Fatalf("expected user denial, got %#v", decision)
	}
}

func TestBrokerAsksBeforeBashTouchesSensitiveConfig(t *testing.T) {
	requests := make(chan *Request, 1)
	broker := NewBroker(requests)
	input := invocation(domain.CategoryCommand, "Bash", map[string]any{"command": "printf '{}' > .codefolio/config.json"})
	result := make(chan domain.PermissionDecision, 1)
	go func() { result <- broker.Authorize(context.Background(), input) }()
	request := <-requests
	if request.Detail != "sensitive Codefolio configuration requires approval" {
		t.Fatalf("unexpected approval detail: %q", request.Detail)
	}
	request.Resolve(Deny)
	if decision := <-result; decision.Reason != "user denied approval" {
		t.Fatalf("expected user denial, got %#v", decision)
	}
}

func TestBrokerPlanAllowsOnlyPlanArtifactWrites(t *testing.T) {
	broker := NewBroker(make(chan *Request, 1))
	plan := filepath.Join(".codefolio", "plans", "session.md")
	allowed := invocation(domain.CategoryWrite, "WriteFile", map[string]any{"file_path": plan})
	allowed.Profile = domain.ProfilePlan
	allowed.PlanFile = filepath.Join(allowed.WorkDir, plan)
	if decision := broker.Authorize(context.Background(), allowed); decision.Effect != domain.PermissionAllow {
		t.Fatalf("expected plan artifact write to be allowed, got %#v", decision)
	}
	blocked := invocation(domain.CategoryWrite, "WriteFile", map[string]any{"file_path": "notes.txt"})
	blocked.Profile = domain.ProfilePlan
	blocked.PlanFile = allowed.PlanFile
	if decision := broker.Authorize(context.Background(), blocked); decision.Effect != domain.PermissionDeny || decision.Reason != "plan profile blocks this action" {
		t.Fatalf("expected plan write denial, got %#v", decision)
	}
}

func TestBrokerPlanBlocksBashAndMCPTools(t *testing.T) {
	broker := NewBroker(make(chan *Request, 1))
	for _, input := range []domain.ToolInvocation{
		invocation(domain.CategoryCommand, "Bash", map[string]any{"command": "go test ./..."}),
		invocation(domain.CategoryCommand, "mcp__github__create_issue", map[string]any{}),
	} {
		input.Profile = domain.ProfilePlan
		if decision := broker.Authorize(context.Background(), input); decision.Effect != domain.PermissionDeny || decision.Reason != "plan profile blocks this action" {
			t.Fatalf("expected plan invocation to be denied, got %#v", decision)
		}
	}
}

func TestBrokerPlanAsksForExternalReads(t *testing.T) {
	requests := make(chan *Request, 1)
	broker := NewBroker(requests)
	input := invocation(domain.CategoryRead, "ReadFile", map[string]any{"file_path": filepath.Join("..", "outside.txt")})
	input.Profile = domain.ProfilePlan
	result := make(chan domain.PermissionDecision, 1)
	go func() { result <- broker.Authorize(context.Background(), input) }()
	request := <-requests
	if request.Detail != "external workspace access requires approval" {
		t.Fatalf("unexpected approval detail: %q", request.Detail)
	}
	request.Resolve(Deny)
	if decision := <-result; decision.Reason != "user denied approval" {
		t.Fatalf("expected user denial, got %#v", decision)
	}
}

func invocation(category domain.ToolCategory, name string, args map[string]any) domain.ToolInvocation {
	workDir, err := filepath.Abs("workspace")
	if err != nil {
		panic(err)
	}
	return domain.ToolInvocation{ID: "tool-1", Name: name, Category: category, Args: args, WorkDir: workDir}
}
