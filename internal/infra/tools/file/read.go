package file

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools/shared"
)

type Reader struct {
	StateCache *StateCache
}

func (t *Reader) Name() string                  { return "ReadFile" }
func (t *Reader) Description() string           { return shared.ReadFileDescription }
func (t *Reader) Category() domain.ToolCategory { return domain.CategoryRead }

func (t *Reader) Schema() map[string]any {
	return map[string]any{
		"name":        t.Name(),
		"description": t.Description(),
		"input_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file_path": map[string]any{"type": "string", "description": "Absolute or relative path to the file to read"},
				"offset":    map[string]any{"type": "integer", "description": "Line offset to start reading from (0-based)", "default": 0},
				"limit":     map[string]any{"type": "integer", "description": "Maximum number of lines to read", "default": 2000},
			},
			"required": []string{"file_path"},
		},
	}
}

func (t *Reader) Execute(_ context.Context, args map[string]any) domain.ToolResult {
	filePath, _ := args["file_path"].(string)
	if filePath == "" {
		return domain.ToolResult{Output: "Error: file_path is required", IsError: true}
	}

	offset := shared.IntArg(args, "offset", 0)
	limit := shared.IntArg(args, "limit", 2000)

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return domain.ToolResult{Output: fmt.Sprintf("Error: file not found: %s", filePath), IsError: true}
	}
	if err != nil {
		return domain.ToolResult{Output: fmt.Sprintf("Error: %s", err), IsError: true}
	}
	if info.IsDir() {
		return domain.ToolResult{Output: fmt.Sprintf("Error: not a file: %s", filePath), IsError: true}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return domain.ToolResult{Output: fmt.Sprintf("Error reading file: %s", err), IsError: true}
	}

	lines := strings.Split(string(data), "\n")
	if offset >= len(lines) {
		return domain.ToolResult{Output: ""}
	}
	end := min(offset+limit, len(lines))
	selected := lines[offset:end]

	if t.StateCache != nil {
		t.StateCache.Record(filePath, string(data), info.ModTime().UnixMilli())
	}

	var sb strings.Builder
	for i, line := range selected {
		if i > 0 {
			sb.WriteByte('\n')
		}
		fmt.Fprintf(&sb, "%d\t%s", i+offset+1, line)
	}
	return domain.ToolResult{Output: sb.String()}
}
