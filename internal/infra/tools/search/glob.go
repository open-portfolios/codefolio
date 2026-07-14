package search

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools/shared"
)

type Glob struct{}

func (t *Glob) Name() string        { return "Glob" }
func (t *Glob) Description() string { return shared.GlobDescription }

func (t *Glob) Category() domain.ToolCategory { return domain.CategoryRead }
func (t *Glob) Schema() map[string]any {
	return map[string]any{
		"name":        t.Name(),
		"description": t.Description(),
		"input_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{"type": "string", "description": "Glob pattern to match (e.g. '**/*.py')"},
				"path":    map[string]any{"type": "string", "description": "Base directory to search from", "default": "."},
			},
			"required": []string{"pattern"},
		},
	}
}

func (t *Glob) Execute(_ context.Context, args map[string]any) domain.ToolResult {
	pattern, _ := args["pattern"].(string)
	basePath, _ := args["path"].(string)
	if basePath == "" {
		basePath = "."
	}
	if pattern == "" {
		return domain.ToolResult{Output: "Error: pattern is required", IsError: true}
	}

	info, err := os.Stat(basePath)
	if os.IsNotExist(err) {
		return domain.ToolResult{Output: fmt.Sprintf("Error: path not found: %s", basePath), IsError: true}
	}
	if err != nil || !info.IsDir() {
		return domain.ToolResult{Output: fmt.Sprintf("Error: path not found: %s", basePath), IsError: true}
	}

	recursive := false
	basePattern := pattern
	for strings.HasPrefix(basePattern, "**/") {
		basePattern = basePattern[3:]
		recursive = true
	}

	var matches []string
	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if shared.SkipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(basePath, path)
		matched := false
		if recursive {
			matched, _ = filepath.Match(basePattern, filepath.Base(path))
		} else {
			matched, _ = filepath.Match(pattern, filepath.Base(path))
			if !matched {
				matched, _ = filepath.Match(pattern, rel)
			}
		}
		if matched {
			matches = append(matches, rel)
		}
		return nil
	})
	if err != nil {
		return domain.ToolResult{Output: fmt.Sprintf("Error: %s", err), IsError: true}
	}

	sort.Strings(matches)
	if len(matches) == 0 {
		return domain.ToolResult{Output: "No files matched the pattern."}
	}
	return domain.ToolResult{Output: strings.Join(matches, "\n")}
}
