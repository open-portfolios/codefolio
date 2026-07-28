package components

import "github.com/cylixlee/tux/renderer"

type SidebarProps struct {
	Messages     int
	InputTokens  int64
	OutputTokens int64
	Status       string
}
type Sidebar struct {
	messages     int
	inputTokens  int64
	outputTokens int64
	status       string
}

func NewSidebar(ctx renderer.Context, props SidebarProps, children ...renderer.Component) *Sidebar {
	return &Sidebar{messages: props.Messages, inputTokens: props.InputTokens, outputTokens: props.OutputTokens, status: props.Status}
}
