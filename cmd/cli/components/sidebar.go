package components

import "github.com/cylixlee/tux/renderer"

type SidebarProps struct {
	WorkDir string
}
type Sidebar struct {
	workDir string
}

func NewSidebar(ctx renderer.Context, props SidebarProps, children ...renderer.Component) *Sidebar {
	return &Sidebar{workDir: props.WorkDir}
}
