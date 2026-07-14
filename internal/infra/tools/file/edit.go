package file

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools/shared"
)

type Editor struct {
	StateCache *StateCache
}

func (t *Editor) Name() string                  { return "EditFile" }
func (t *Editor) Description() string           { return shared.EditFileDescription }
func (t *Editor) Category() domain.ToolCategory { return domain.CategoryWrite }

func (t *Editor) Schema() map[string]any {
	return map[string]any{
		"name":        t.Name(),
		"description": t.Description(),
		"input_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file_path":  map[string]any{"type": "string", "description": "Path to the file to edit"},
				"old_string": map[string]any{"type": "string", "description": "The exact string to find and replace (must be unique in file)"},
				"new_string": map[string]any{"type": "string", "description": "The replacement string"},
			},
			"required": []string{"file_path", "old_string", "new_string"},
		},
	}
}

func (t *Editor) Execute(_ context.Context, args map[string]any) domain.ToolResult {
	filePath, _ := args["file_path"].(string)
	oldStr, _ := args["old_string"].(string)
	newStr, _ := args["new_string"].(string)

	if filePath == "" {
		return domain.ToolResult{Output: "Error: file_path is required", IsError: true}
	}

	if t.StateCache != nil {
		if ok, errMsg := t.StateCache.Check(filePath); !ok {
			return domain.ToolResult{Output: errMsg, IsError: true}
		}
	}

	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return domain.ToolResult{Output: fmt.Sprintf("Error: file not found: %s", filePath), IsError: true}
	}
	if err != nil {
		return domain.ToolResult{Output: fmt.Sprintf("Error reading file: %s", err), IsError: true}
	}

	content := string(data)
	count := strings.Count(content, oldStr)
	if count == 0 {
		return domain.ToolResult{Output: "Error: old_string not found in file", IsError: true}
	}
	if count > 1 {
		return domain.ToolResult{Output: fmt.Sprintf("Error: old_string found %d times, must be unique", count), IsError: true}
	}

	newContent := strings.Replace(content, oldStr, newStr, 1)
	if err := os.WriteFile(filePath, []byte(newContent), 0o644); err != nil {
		return domain.ToolResult{Output: fmt.Sprintf("Error writing file: %s", err), IsError: true}
	}

	if t.StateCache != nil {
		t.StateCache.Update(filePath, newContent)
	}

	return domain.ToolResult{Output: fmt.Sprintf("Successfully edited %s", filePath)}
}
