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
		return builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "shell-content", Gap: 1}, s.children...)
	}
	width, _ := ctx.Size()
	main := builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "session-main", Flex: 1},
		builtin.CreateText(ctx, builtin.TextProps{Key: "timeline-top-space", Text: "", Bg: Theme.Background}),
		horizontalInset(ctx, "timeline-inset", s.children[0], true),
		builtin.CreateText(ctx, builtin.TextProps{Key: "timeline-bottom-space", Text: "", Bg: Theme.Background}),
		horizontalInset(ctx, "prompt-inset", s.children[2], false),
		s.children[3],
		s.children[4],
	)
	if width > 120 {
		return builtin.CreateRow(ctx, builtin.RowProps{Key: "session", Flex: 1}, main, s.children[1])
	}
	return main.Render(ctx)
}

func horizontalInset(ctx renderer.Context, key string, child renderer.Component, flex bool) *renderer.Element {
	e := &renderer.Element{}
	e.SetKey(key)
	e.SetTag("horizontal-inset")
	e.SetChildren(renderer.RenderChildren(ctx, child))
	if flex {
		e.SetFlex(1)
	}
	for _, element := range e.Children() {
		e.SetLayoutWidth(max(e.LayoutWidth(), element.LayoutWidth()+4))
		e.SetLayoutHeight(max(e.LayoutHeight(), element.LayoutHeight()))
	}
	e.SetPaintFn(func(draw renderer.DrawContext, box renderer.Rect) {
		for _, element := range e.Children() {
			element.Paint(draw.Viewport(box).Viewport(renderer.Rect{X: 2, Width: max(box.Width-4, 0), Height: box.Height}), renderer.Rect{Width: max(box.Width-4, 0), Height: box.Height})
		}
		e.SetRect(renderer.Rect{X: draw.Origin.X + box.X, Y: draw.Origin.Y + box.Y, Width: box.Width, Height: box.Height})
	})
	return e
}
