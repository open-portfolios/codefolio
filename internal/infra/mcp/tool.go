package mcp

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools/shared"
)

var nonToolName = regexp.MustCompile(`[^A-Za-z0-9_]`)

type Tool struct {
	serverName string
	remote     RemoteTool
	client     Client
}

func NewTool(serverName string, remote RemoteTool, client Client) *Tool {
	return &Tool{serverName: serverName, remote: remote, client: client}
}

func (t *Tool) Name() string {
	return "mcp__" + sanitizeName(t.serverName) + "__" + sanitizeName(t.remote.Name)
}

func (t *Tool) Description() string {
	if t.remote.Description == "" {
		return "MCP tool from " + t.serverName
	}
	return t.remote.Description
}

func (t *Tool) Category() domain.ToolCategory { return domain.CategoryCommand }

func (t *Tool) ShouldDefer() bool { return true }

func (t *Tool) Schema() map[string]any {
	return map[string]any{
		"name":         t.Name(),
		"description":  t.Description(),
		"input_schema": t.remote.InputSchema,
	}
}

func (t *Tool) Execute(ctx context.Context, args map[string]any) domain.ToolResult {
	if err := ctx.Err(); err != nil {
		return domain.ToolResult{Output: "Error: MCP tool cancelled: " + err.Error(), IsError: true, Outcome: domain.ToolOutcomeCancelled}
	}
	result, err := t.client.CallTool(ctx, t.remote.Name, args)
	if err != nil {
		return domain.ToolResult{Output: fmt.Sprintf("Error: MCP tool %s failed: %v", t.Name(), err), IsError: true, Outcome: domain.ToolOutcomeFailed}
	}
	output := truncate(result.Output)
	if result.IsError {
		if !strings.HasPrefix(output, "Error:") {
			output = "Error: " + output
		}
		return domain.ToolResult{Output: output, IsError: true, Outcome: domain.ToolOutcomeFailed}
	}
	return domain.ToolResult{Output: output, Outcome: domain.ToolOutcomeSucceeded}
}

func sanitizeName(name string) string { return nonToolName.ReplaceAllString(name, "_") }

func truncate(output string) string {
	if len(output) <= shared.MaxOutputChars {
		return output
	}
	return output[:shared.MaxOutputChars] + "\n… output truncated"
}

var _ domain.DeferrableTool = (*Tool)(nil)
