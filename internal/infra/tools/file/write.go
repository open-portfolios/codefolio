package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools/shared"
)

type Writer struct {
	StateCache *StateCache
}

func (t *Writer) Name() string                  { return "WriteFile" }
func (t *Writer) Description() string           { return shared.WriteFileDescription }
func (t *Writer) Category() domain.ToolCategory { return domain.CategoryWrite }

func (t *Writer) Schema() map[string]any {
	return map[string]any{
		"name":        t.Name(),
		"description": t.Description(),
		"input_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file_path": map[string]any{"type": "string", "description": "Path to the file to write"},
				"content":   map[string]any{"type": "string", "description": "Content to write to the file"},
			},
			"required": []string{"file_path", "content"},
		},
	}
}

func (t *Writer) Execute(_ context.Context, args map[string]any) domain.ToolResult {
	filePath, _ := args["file_path"].(string)
	content, _ := args["content"].(string)
	if filePath == "" {
		return domain.ToolResult{Output: "Error: file_path is required", IsError: true}
	}

	if t.StateCache != nil {
		if _, err := os.Stat(filePath); err == nil {
			if ok, errMsg := t.StateCache.Check(filePath); !ok {
				return domain.ToolResult{Output: errMsg, IsError: true}
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return domain.ToolResult{Output: fmt.Sprintf("Error creating directories: %s", err), IsError: true}
	}

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return domain.ToolResult{Output: fmt.Sprintf("Error writing file: %s", err), IsError: true}
	}

	if t.StateCache != nil {
		t.StateCache.Update(filePath, content)
	}

	return domain.ToolResult{Output: fmt.Sprintf("Successfully wrote to %s", filePath)}
}
