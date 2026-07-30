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

func TestBrokerAsksForNonRootRemoval(t *testing.T) {
	requests := make(chan *Request, 1)
	broker := NewBroker(requests)
	result := make(chan domain.PermissionDecision, 1)
	go func() {
		result <- broker.Authorize(context.Background(), invocation(domain.CategoryCommand, "Bash", map[string]any{"command": "rm -rf /tmp/build"}))
	}()
	request := <-requests
	request.Resolve(Deny)
	if decision := <-result; decision.Reason != "user denied approval" {
		t.Fatalf("expected user approval for non-root removal, got %#v", decision)
	}
}

func TestBrokerRequiresApprovalAndRemembersSessionGrant(t *testing.T) {
	requests := make(chan *Request, 1)
	broker := NewBroker(requests)
	invocation := invocation(domain.CategoryWrite, "WriteFile", map[string]any{"file_path": "notes.txt"})

	result := make(chan domain.PermissionDecision, 1)
	go func() { result <- broker.Authorize(context.Background(), invocation) }()
	request := <-requests
	request.Resolve(AllowSession)
	if decision := <-result; decision.Effect != domain.PermissionAllow {
		t.Fatalf("expected approval, got %#v", decision)
	}
	if decision := broker.Authorize(context.Background(), invocation); decision.Effect != domain.PermissionAllow || decision.Reason != "allowed for this session" {
		t.Fatalf("expected session grant, got %#v", decision)
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

func TestBrokerDeniesProtectedConfig(t *testing.T) {
	broker := NewBroker(make(chan *Request, 1))
	decision := broker.Authorize(context.Background(), invocation(domain.CategoryWrite, "WriteFile", map[string]any{"file_path": filepath.Join(".codefolio", "config.json")}))
	if decision.Effect != domain.PermissionDeny || decision.Reason != "protected Codefolio configuration" {
		t.Fatalf("expected protected path denial, got %#v", decision)
	}
}

func invocation(category domain.ToolCategory, name string, args map[string]any) domain.ToolInvocation {
	workDir, err := filepath.Abs("workspace")
	if err != nil {
		panic(err)
	}
	return domain.ToolInvocation{ID: "tool-1", Name: name, Category: category, Args: args, WorkDir: workDir}
}
