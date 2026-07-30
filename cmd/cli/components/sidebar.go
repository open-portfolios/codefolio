package components

import (
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/style"
)

type SidebarProps struct {
	WorkDir string
	Mcp     MCPStatus
}

type MCPStatus struct {
	Label string
	Color style.Color
}

type Sidebar struct {
	workDir string
	mcp     MCPStatus
}

func NewSidebar(ctx renderer.Context, props SidebarProps, children ...renderer.Component) *Sidebar {
	return &Sidebar{workDir: props.WorkDir, mcp: props.Mcp}
}

func (s *Sidebar) MCPColor() style.Color { return s.mcp.Color }
