package toolsearch

import (
	"context"
	"testing"

	"github.com/open-portfolios/codefolio/internal/domain"
)

func TestToolSearchLoadsDeferredTool(t *testing.T) {
	registry := &fakeRegistry{tools: map[string]domain.Tool{"mcp__github__search": &fakeTool{name: "mcp__github__search", deferred: true}}}
	tool := &ToolSearch{Registry: registry}
	result := tool.Execute(context.Background(), map[string]any{"query": "github"})
	if result.IsError {
		t.Fatalf("unexpected search result: %#v", result)
	}
	if !registry.discovered["mcp__github__search"] {
		t.Fatalf("tool was not marked discovered: %#v", registry.discovered)
	}
}

type fakeRegistry struct {
	tools      map[string]domain.Tool
	discovered map[string]bool
}

func (r *fakeRegistry) Register(tool domain.Tool) error {
	if r.tools == nil {
		r.tools = make(map[string]domain.Tool)
	}
	r.tools[tool.Name()] = tool
	return nil
}

func (r *fakeRegistry) MarkDiscovered(name string) {
	if r.discovered == nil {
		r.discovered = make(map[string]bool)
	}
	r.discovered[name] = true
}

func (r *fakeRegistry) FindDeferredByNames(names []string) []map[string]any {
	var result []map[string]any
	for _, name := range names {
		if tool := r.tools[name]; tool != nil {
			result = append(result, tool.Schema())
		}
	}
	return result
}

func (r *fakeRegistry) SearchDeferred(query string, maxResults int) []map[string]any {
	var result []map[string]any
	for _, tool := range r.tools {
		if len(result) == maxResults {
			break
		}
		result = append(result, tool.Schema())
	}
	return result
}

func (r *fakeRegistry) GetDeferredToolNames() []string { return nil }

type fakeTool struct {
	name     string
	deferred bool
}

func (t *fakeTool) Name() string                  { return t.name }
func (t *fakeTool) Description() string           { return "search" }
func (t *fakeTool) Category() domain.ToolCategory { return domain.CategoryCommand }
func (t *fakeTool) ShouldDefer() bool             { return t.deferred }
func (t *fakeTool) Schema() map[string]any {
	return map[string]any{"name": t.name, "description": t.Description(), "input_schema": map[string]any{"type": "object"}}
}
func (t *fakeTool) Execute(context.Context, map[string]any) domain.ToolResult {
	return domain.ToolResult{}
}
