package search

import (
	"os"
	"path/filepath"
	"testing"
)

func setupGlobTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := []string{
		"main.go",
		"cmd/cli/main.go",
		"internal/agents/agent.go",
		"internal/agents/agent_test.go",
		"docs/readme.md",
	}
	for _, f := range files {
		fp := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fp, []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestGlobDoubleStarPattern(t *testing.T) {
	dir := setupGlobTree(t)
	g := &Glob{}
	result := g.Execute(nil, map[string]any{"pattern": "**/*.go", "path": dir})
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if result.Output == "No files matched the pattern." {
		t.Fatal("expected matches for **/*.go")
	}
}

func TestGlobPlainPattern(t *testing.T) {
	dir := setupGlobTree(t)
	g := &Glob{}
	result := g.Execute(nil, map[string]any{"pattern": "*.go", "path": dir})
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if result.Output == "No files matched the pattern." {
		t.Fatal("expected matches for *.go at root")
	}
}

func TestGlobNoMatch(t *testing.T) {
	dir := setupGlobTree(t)
	g := &Glob{}
	result := g.Execute(nil, map[string]any{"pattern": "*.py", "path": dir})
	if result.IsError {
		t.Fatal("expected no error for no-match pattern")
	}
	if result.Output != "No files matched the pattern." {
		t.Errorf("expected 'No files matched', got %q", result.Output)
	}
}

func TestGlobMissingPath(t *testing.T) {
	g := &Glob{}
	result := g.Execute(nil, map[string]any{"pattern": "*.go", "path": "/nonexistent/path"})
	if !result.IsError {
		t.Fatal("expected error for missing path")
	}
}
