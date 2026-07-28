package components

import (
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/renderer"
)

type ShellProps struct{}
type Shell struct{ children []renderer.Component }

func NewShell(ctx renderer.Context, props ShellProps, children ...renderer.Component) *Shell {
	return &Shell{children: children}
}

func (s *Shell) Render(ctx renderer.Context) *renderer.Element {
	if len(s.children) != 5 {
		return builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "root", Padding: 1, Gap: 1}, s.children...)
	}
	width, _ := ctx.Size()
	mainChildren := []renderer.Component{s.children[1]}
	if width >= 120 {
		mainChildren = append(mainChildren, s.children[2])
	}
	main := builtin.CreateRow(ctx, builtin.RowProps{Key: "main", Flex: 1, Gap: 1}, mainChildren...)
	return builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "root", Padding: 1, Gap: 1}, s.children[0], main, s.children[3], s.children[4])
}
