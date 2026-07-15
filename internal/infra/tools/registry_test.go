package tools

import (
	"context"
	"testing"

	"github.com/open-portfolios/codefolio/internal/domain"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	r := newRegistry()
	r.Register(&stubTool{name: "TestTool"})
	tool := r.Get("TestTool")
	if tool == nil {
		t.Fatal("expected tool to be registered")
	}
	if tool.Name() != "TestTool" {
		t.Errorf("expected name 'TestTool', got %q", tool.Name())
	}
}

func TestRegistryGetAllSchemasIncludesNonDeferred(t *testing.T) {
	r := newRegistry()
	r.Register(&stubTool{name: "A"})
	r.Register(&stubTool{name: "B"})
	schemas := r.GetAllSchemas()
	if len(schemas) != 2 {
		t.Errorf("expected 2 schemas, got %d", len(schemas))
	}
}

func TestRegistryDeferredToolExcluded(t *testing.T) {
	r := newRegistry()
	r.Register(&stubTool{name: "A"})
	r.Register(&stubDeferredTool{stubTool{name: "B"}})
	schemas := r.GetAllSchemas()
	if len(schemas) != 1 {
		t.Errorf("expected 1 schema (deferred excluded), got %d", len(schemas))
	}
}

func TestRegistryDeferredToolDiscovered(t *testing.T) {
	r := newRegistry()
	r.Register(&stubDeferredTool{stubTool{name: "B"}})
	r.MarkDiscovered("B")
	schemas := r.GetAllSchemas()
	if len(schemas) != 1 {
		t.Errorf("expected 1 schema after discovery, got %d", len(schemas))
	}
}

func TestRegistryOpenAIFormat(t *testing.T) {
	r := newRegistry()
	r.protocol = "openai"
	r.Register(&stubTool{name: "A"})
	schemas := r.GetAllSchemas()
	if len(schemas) != 1 {
		t.Fatal("expected 1 schema")
	}
	s := schemas[0]
	if s["type"] != "function" {
		t.Errorf("expected type 'function', got %v", s["type"])
	}
	if s["parameters"] == nil {
		t.Error("expected 'parameters' key in openai format")
	}
}

func TestRegistryGetDeferredToolNames(t *testing.T) {
	r := newRegistry()
	r.Register(&stubTool{name: "A"})
	r.Register(&stubDeferredTool{stubTool{name: "Hidden"}})
	names := r.GetDeferredToolNames()
	if len(names) != 1 || names[0] != "Hidden" {
		t.Errorf("expected ['Hidden'], got %v", names)
	}
}

type stubTool struct{ name string }

func (s *stubTool) Name() string                  { return s.name }
func (s *stubTool) Description() string           { return "A test tool" }
func (s *stubTool) Category() domain.ToolCategory { return domain.CategoryRead }
func (s *stubTool) Schema() map[string]any {
	return map[string]any{
		"name":        s.name,
		"description": "A test tool",
		"input_schema": map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
}
func (s *stubTool) Execute(_ context.Context, _ map[string]any) domain.ToolResult {
	return domain.ToolResult{Output: "ok"}
}

type stubDeferredTool struct{ stubTool }

func (s *stubDeferredTool) ShouldDefer() bool { return true }
