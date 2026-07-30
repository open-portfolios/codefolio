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

func TestRegistryReturnsCanonicalSchema(t *testing.T) {
	r := newRegistry()
	r.Register(&stubTool{name: "A"})
	schemas := r.GetAllSchemas()
	if len(schemas) != 1 {
		t.Fatal("expected 1 schema")
	}
	s := schemas[0]
	if s["name"] != "A" {
		t.Errorf("expected name 'A', got %v", s["name"])
	}
	if s["input_schema"] == nil {
		t.Error("expected canonical 'input_schema' key")
	}
	if _, ok := s["parameters"]; ok {
		t.Error("registry must not return provider-specific 'parameters' key")
	}
}

func TestRegistryDeferredSchemasRemainCanonical(t *testing.T) {
	r := newRegistry()
	r.Register(&stubDeferredTool{stubTool{name: "Hidden"}})

	schemas := r.FindDeferredByNames([]string{"Hidden"})
	if len(schemas) != 1 {
		t.Fatal("expected one deferred schema")
	}
	if schemas[0]["input_schema"] == nil {
		t.Error("expected canonical 'input_schema' key")
	}
	if _, ok := schemas[0]["parameters"]; ok {
		t.Error("deferred schema must not contain provider-specific 'parameters' key")
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

func TestRegistryRejectsDuplicateTools(t *testing.T) {
	r := newRegistry()
	if err := r.Register(&stubTool{name: "A"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(&stubTool{name: "A"}); err == nil {
		t.Fatal("expected duplicate tool error")
	}
}

func TestRegistryReturnsSchemasInNameOrder(t *testing.T) {
	r := newRegistry()
	_ = r.Register(&stubTool{name: "B"})
	_ = r.Register(&stubTool{name: "A"})
	schemas := r.GetAllSchemas()
	if schemas[0]["name"] != "A" || schemas[1]["name"] != "B" {
		t.Fatalf("schemas are not stable: %#v", schemas)
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
