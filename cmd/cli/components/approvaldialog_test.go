package components

import (
	"testing"

	"github.com/open-portfolios/codefolio/internal/infra/approval"
)

func TestDisplayToolName(t *testing.T) {
	if got := displayToolName("mcp__chrome_devtools__list_pages"); got != "MCP chrome_devtools / list_pages" {
		t.Fatalf("MCP display name = %q", got)
	}
	if got := displayToolName("Bash"); got != "Bash" {
		t.Fatalf("regular display name = %q", got)
	}
}

func TestApprovalSummaryUsesMCPDisplayName(t *testing.T) {
	request := &approval.Request{ToolName: "mcp__chrome_devtools__new_page", Summary: "mcp__chrome_devtools__new_page"}
	toolName, summary := displayToolName(request.ToolName), request.Summary
	if summary == request.ToolName {
		summary = toolName
	}
	if summary != "MCP chrome_devtools / new_page" {
		t.Fatalf("approval summary = %q", summary)
	}
}
