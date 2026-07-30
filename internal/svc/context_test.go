package svc

import (
	"context"
	"strings"
	"testing"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

type summaryDriver struct{}

func (summaryDriver) Stream(ctx context.Context, messages []llm.Message, options ...llm.StreamOption) (<-chan llm.Delta, <-chan error) {
	deltas := make(chan llm.Delta, 1)
	errs := make(chan error)
	deltas <- llm.MessageDelta{Content: "Summary of earlier work."}
	close(deltas)
	close(errs)
	return deltas, errs
}

func TestPreviewToolOutputsKeepsLedgerUntouched(t *testing.T) {
	original := strings.Repeat("x", toolResultTokenLimit*3+1)
	messages := []domain.ChatMessage{{Role: llm.RoleTool, Content: original, ToolCallID: "call-1"}}
	working, trimmed := previewToolOutputs(messages)
	if trimmed != 1 {
		t.Fatalf("trimmed = %d, want 1", trimmed)
	}
	if messages[0].Content != original {
		t.Fatal("preview mutated the session ledger")
	}
	if !strings.Contains(working[0].Content, "Tool output truncated") {
		t.Fatalf("working preview = %q, want truncation marker", working[0].Content)
	}
	if working[0].ToolCallID != "call-1" {
		t.Fatal("preview lost the tool-call relationship")
	}
}

func TestContextManagerCompactsOnlyCompletedTurns(t *testing.T) {
	manager := NewContextManager(summaryDriver{})
	messages := []domain.ChatMessage{
		{ID: "system", Role: llm.RoleSystem, Content: "policy"},
		{ID: "u-1", Role: llm.RoleUser, Content: strings.Repeat("a", 100)},
		{ID: "a-1", Role: llm.RoleAssistant, ToolCalls: []domain.ToolCall{{ID: "call-1", Name: "ReadFile", Input: `{}`}}},
		{ID: "t-1", Role: llm.RoleTool, ToolCallID: "call-1", Content: "old tool result"},
		{ID: "u-2", Role: llm.RoleUser, Content: strings.Repeat("b", 100)},
		{ID: "u-3", Role: llm.RoleUser, Content: strings.Repeat("c", 100)},
		{ID: "u-4", Role: llm.RoleUser, Content: strings.Repeat("d", 100)},
	}
	prepared, err := manager.Prepare(context.Background(), messages, nil, &conf.Struct{Context: conf.Context{WindowTokens: 300, MaxOutputTokens: 50}})
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	if len(prepared.Messages) != 5 {
		t.Fatalf("prepared messages = %d, want system + summary + 3 retained turns", len(prepared.Messages))
	}
	if prepared.Messages[1].ID != "context-summary" {
		t.Fatalf("message[1].ID = %q, want context-summary", prepared.Messages[1].ID)
	}
	for _, message := range prepared.Messages {
		if message.ID == "a-1" || message.ID == "t-1" {
			t.Fatalf("compacted context retained only part of the oldest tool round: %#v", message)
		}
	}
	if messages[2].ToolCalls[0].ID != "call-1" || messages[3].Content != "old tool result" {
		t.Fatal("compaction mutated the original transcript")
	}
}
