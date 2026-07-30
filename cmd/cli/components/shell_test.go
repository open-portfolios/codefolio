package components

import (
	"testing"

	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/renderer"
)

type shellTestComponent struct{}

func (shellTestComponent) Render(ctx renderer.Context) *renderer.Element {
	return builtin.CreateText(ctx, builtin.TextProps{Text: "content", Fg: Theme.Text})
}

func TestShellShowsSidebarOnlyAboveOpenCodeBreakpoint(t *testing.T) {
	for _, test := range []struct {
		width       int
		wantSidebar bool
	}{
		{width: 120, wantSidebar: false},
		{width: 121, wantSidebar: true},
	} {
		ctx := renderer.Context{SizeFn: func() (int, int) { return test.width, 30 }}
		shell := NewShell(ctx, ShellProps{}, shellTestComponent{}, shellTestComponent{}, shellTestComponent{}, shellTestComponent{}, shellTestComponent{})
		root := shell.Render(ctx)
		if got := len(root.Children()) == 2; got != test.wantSidebar {
			t.Errorf("width %d sidebar visible = %t, want %t", test.width, got, test.wantSidebar)
		}
	}
}
