package chat

import "testing"

func TestBuildToolsUsesCanonicalInputSchema(t *testing.T) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string"},
		},
		"required": []string{"path"},
	}

	tools, err := buildTools([]map[string]any{{
		"name":         "ReadFile",
		"description":  "Read a file.",
		"input_schema": inputSchema,
	}})
	if err != nil {
		t.Fatalf("buildTools returned error: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected one tool, got %d", len(tools))
	}

	function := tools[0].OfFunction
	if function == nil {
		t.Fatal("expected function tool")
	}
	if function.Function.Name != "ReadFile" {
		t.Errorf("expected name ReadFile, got %q", function.Function.Name)
	}
	if got := map[string]any(function.Function.Parameters); got["type"] != "object" {
		t.Errorf("expected object parameters, got %v", got["type"])
	}
	if got := map[string]any(function.Function.Parameters); got["properties"] == nil {
		t.Error("expected parameters to include properties")
	}
}

func TestBuildToolsRejectsMissingInputSchema(t *testing.T) {
	_, err := buildTools([]map[string]any{{"name": "ReadFile"}})
	if err == nil {
		t.Fatal("expected an error for a missing input_schema")
	}
}
