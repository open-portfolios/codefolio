package svc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/open-portfolios/codefolio/internal/domain"
)

func TestLoadAgentsCombinesAncestorFiles(t *testing.T) {
	root := t.TempDir()
	workDir := filepath.Join(root, "nested", "project")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("root instruction"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "nested", "AGENTS.md"), []byte("nested instruction"), 0o600); err != nil {
		t.Fatal(err)
	}
	content := loadAgents(root, workDir)
	if !strings.Contains(content, "root instruction") || !strings.Contains(content, "nested instruction") {
		t.Fatalf("expected ancestor instructions, got %q", content)
	}
	if strings.Index(content, "root instruction") > strings.Index(content, "nested instruction") {
		t.Fatalf("expected root instruction before nested instruction, got %q", content)
	}
}

func TestSaveMemoryRoutesTopicAndRebuildsIndex(t *testing.T) {
	root := t.TempDir()
	proposal := memoryProposal{
		Type:        "project",
		Name:        "Release Process",
		Description: "The production release constraint.",
		Content:     "Production releases require an approved change window.",
	}
	if err := saveMemory(root, proposal); err != nil {
		t.Fatal(err)
	}
	topic := readLimited(filepath.Join(root, "release-process.md"), memoryTopicMaxBytes)
	if !strings.Contains(topic, "type: project") || !strings.Contains(topic, proposal.Content) {
		t.Fatalf("unexpected topic content: %q", topic)
	}
	index := readLimited(filepath.Join(root, "MEMORY.md"), memoryIndexMaxBytes)
	if !strings.Contains(index, "release-process.md") || !strings.Contains(index, proposal.Description) {
		t.Fatalf("unexpected index content: %q", index)
	}
}

func TestValidProposalRejectsSecretsAndUnknownTypes(t *testing.T) {
	valid := memoryProposal{Type: "user", Name: "Style", Description: "Writing style", Content: "Prefer concise answers."}
	if !validProposal(valid) {
		t.Fatal("expected valid proposal")
	}
	valid.Content = "The api key is abc"
	if validProposal(valid) {
		t.Fatal("expected proposal containing a secret marker to be rejected")
	}
	valid.Content = "Prefer concise answers."
	valid.Type = "unknown"
	if validProposal(valid) {
		t.Fatal("expected unknown type to be rejected")
	}
}

func TestExtractionPromptUsesOnlyTheLastUserTurn(t *testing.T) {
	messages := []domain.ChatMessage{
		{Role: "user", Content: "old request"},
		{Role: "assistant", Content: "old answer"},
		{Role: "user", Content: "remember this preference"},
		{Role: "assistant", Content: "I will use concise summaries."},
	}
	prompt := extractionPrompt(messages)
	if strings.Contains(prompt, "old request") || strings.Contains(prompt, "old answer") {
		t.Fatalf("extraction prompt included an earlier turn: %q", prompt)
	}
	if !strings.Contains(prompt, "remember this preference") {
		t.Fatalf("extraction prompt omitted the completed turn: %q", prompt)
	}
}
