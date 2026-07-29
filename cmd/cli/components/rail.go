package components

import (
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/style"
)

type RailProps struct{}
type Rail struct{ children []renderer.Component }

func NewRail(ctx renderer.Context, props RailProps, children ...renderer.Component) *Rail {
	return &Rail{children: children}
}

func (r *Rail) Render(ctx renderer.Context) *renderer.Element {
	e := &renderer.Element{}
	e.SetKey("composer-rail")
	e.SetTag("composer-rail")
	e.SetChildren(renderer.RenderChildren(ctx, r.children...))
	for _, child := range e.Children() {
		e.SetLayoutWidth(max(e.LayoutWidth(), child.LayoutWidth()+1))
		e.SetLayoutHeight(max(e.LayoutHeight(), child.LayoutHeight()))
	}
	e.SetPaintFn(func(draw renderer.DrawContext, box renderer.Rect) {
		draw.Fill(box, renderer.Cell{Rune: ' ', Style: style.Style{Bg: Theme.BackgroundElement}})
		for row := range box.Height {
			draw.WriteStringWide(box.X, box.Y+row, "┃", style.Style{Fg: Theme.Primary, Bg: Theme.BackgroundElement, Attrs: style.Bold})
		}
		for _, child := range e.Children() {
			child.Paint(draw.Viewport(box).Viewport(renderer.Rect{X: 1, Width: max(box.Width-1, 0), Height: box.Height}), renderer.Rect{Width: max(box.Width-1, 0), Height: box.Height})
		}
		e.SetRect(renderer.Rect{X: draw.Origin.X + box.X, Y: draw.Origin.Y + box.Y, Width: box.Width, Height: box.Height})
	})
	return e
}
