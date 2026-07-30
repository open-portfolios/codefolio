package main

import (
	"testing"

	"github.com/open-portfolios/codefolio/cmd/cli/components"
	"github.com/open-portfolios/codefolio/internal/infra/mcp"
)

func TestMCPStatusFor(t *testing.T) {
	tests := []struct {
		name       string
		summary    mcp.Summary
		connecting bool
		want       string
	}{
		{name: "none configured", want: "○ No servers configured"},
		{name: "connecting", summary: mcp.Summary{Configured: 2}, connecting: true, want: "◌ Connecting 2 server(s)..."},
		{name: "ready", summary: mcp.Summary{Configured: 2, Ready: 2, Tools: 5}, want: "● 2 server(s) · 5 tools"},
		{name: "partial", summary: mcp.Summary{Configured: 2, Ready: 1, Unavailable: 1}, want: "▲ 1 ready · 1 unavailable"},
		{name: "unavailable", summary: mcp.Summary{Configured: 1, Unavailable: 1}, want: "× 1 server(s) unavailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := mcpStatusFor(tt.summary, tt.connecting)
			if status.Label != tt.want {
				t.Fatalf("label = %q, want %q", status.Label, tt.want)
			}
		})
	}

	if status := mcpStatusFor(mcp.Summary{Configured: 1}, false); status.Color != components.Theme.Error {
		t.Fatal("pending server should not be presented as ready")
	}
}
