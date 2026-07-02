package builtin

import (
	"errors"
	"testing"

	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/debug"
)

func TestContainerRenderEmpty(t *testing.T) {
	renderer := debug.NewRenderer(1, 1)
	component := Container(ContainerProps{})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}
}

func TestContainerRenderChildrenInOrder(t *testing.T) {
	renderer := debug.NewRenderer(1, 3)
	component := Container(
		ContainerProps{},
		paintComponent{column: 0, b: 'a'},
		paintComponent{column: 1, b: 'b'},
		paintComponent{column: 2, b: 'c'},
	)

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.String(0, 0, 3), "abc"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestContainerRenderBuildsCompositeChildren(t *testing.T) {
	renderer := debug.NewRenderer(1, 1)
	component := Container(
		ContainerProps{},
		compositeComponent{child: paintComponent{b: 'x'}},
	)

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.Buf[0][0].Ch, rune('x'); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestContainerRenderReturnsChildError(t *testing.T) {
	renderer := debug.NewRenderer(1, 1)
	want := errors.New("render failed")
	component := Container(ContainerProps{}, errorComponent{err: want})

	if got := component.Render(tux.BuildContext{}, renderer); !errors.Is(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

type paintComponent struct {
	tux.Atomic

	column int
	b      rune
}

func (c paintComponent) Render(_ tux.BuildContext, ctx tux.RenderContext) error {
	ctx.Paint(0, c.column, tux.Cell{Ch: c.b, Width: 1})
	return nil
}

type compositeComponent struct {
	tux.Composite

	child tux.Component
}

func (c compositeComponent) Build(tux.BuildContext) tux.Component {
	return c.child
}

type errorComponent struct {
	tux.Atomic

	err error
}

func (c errorComponent) Render(tux.BuildContext, tux.RenderContext) error {
	return c.err
}
