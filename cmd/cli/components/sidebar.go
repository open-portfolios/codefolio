package components

import (
	"fmt"

	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/style"
	"github.com/open-portfolios/codefolio/internal/domain"
)

type SidebarProps struct {
	WorkDir string
	Mcp     MCPStatus
	Context domain.ContextMetrics
}

type MCPStatus struct {
	Label string
	Color style.Color
}

type Sidebar struct {
	workDir string
	mcp     MCPStatus
	context domain.ContextMetrics
}

func NewSidebar(ctx renderer.Context, props SidebarProps, children ...renderer.Component) *Sidebar {
	return &Sidebar{workDir: props.WorkDir, mcp: props.Mcp, context: props.Context}
}

func (s *Sidebar) MCPColor() style.Color { return s.mcp.Color }

func (s *Sidebar) ContextTokens() string {
	used, usable := s.context.UsedInputTokens(), s.context.UsableInputTokens()
	if usable == 0 {
		return "Context profile unavailable"
	}
	return fmt.Sprintf("%s / %s tokens", formatTokens(used), formatTokens(usable))
}

func (s *Sidebar) ContextPercent() string {
	if s.context.UsableInputTokens() == 0 {
		return "Awaiting context profile"
	}
	return fmt.Sprintf("%d%% used · %s", s.context.UsagePercent(), contextSource(s.context))
}

func (s *Sidebar) ContextReserve() string {
	return fmt.Sprintf("%s output reserve", formatTokens(s.context.ReservedOutputTokens))
}

func contextSource(metrics domain.ContextMetrics) string {
	if metrics.ActualInputTokens > 0 {
		return "provider measured"
	}
	return "estimated"
}

func formatTokens(tokens int64) string {
	if tokens < 1_000 {
		return fmt.Sprintf("%d", tokens)
	}
	return fmt.Sprintf("%.1fk", float64(tokens)/1_000)
}
