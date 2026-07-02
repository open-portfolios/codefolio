package debug

import (
	"testing"

	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/misc"
)

func TestNewRendererCreatesBuffer(t *testing.T) {
	renderer := NewRenderer(2, 3)

	if len(renderer.Buf) != 2 {
		t.Fatalf("got %d rows, want 2", len(renderer.Buf))
	}
	for row := range renderer.Buf {
		if len(renderer.Buf[row]) != 3 {
			t.Fatalf("row %d has %d columns, want 3", row, len(renderer.Buf[row]))
		}
	}
}

func TestRendererPaintWritesCell(t *testing.T) {
	renderer := NewRenderer(2, 3)
	cell := tux.Cell{Ch: 'x', Width: 1, Foreground: misc.ColorRed, Background: misc.ColorBlue}

	renderer.Paint(1, 2, cell)

	if got, want := renderer.Buf[1][2], cell; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRendererSubmit(t *testing.T) {
	renderer := NewRenderer(1, 1)

	if err := renderer.Submit(); err != nil {
		t.Fatal(err)
	}
}

func TestRendererRender(t *testing.T) {
	renderer := NewRenderer(1, 1)

	if err := renderer.Render(tux.BuildContext{}, paintComponent{b: 'x'}); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.Buf[0][0].Ch, rune('x'); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRendererString(t *testing.T) {
	renderer := NewRenderer(1, 4)
	renderer.Paint(0, 0, tux.Cell{Ch: 'a', Width: 1})
	renderer.Paint(0, 1, tux.Cell{Ch: 'b', Width: 1})
	renderer.Paint(0, 2, tux.Cell{Ch: 'c', Width: 1})
	renderer.Paint(0, 3, tux.Cell{Ch: 'd', Width: 1})

	if got, want := renderer.String(0, 1, 3), "bc"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRendererQueueCursorMove(t *testing.T) {
	renderer := NewRenderer(2, 3)

	renderer.QueueCursorMove(1, 2)

	row, column, queued := renderer.CursorPos()
	if !queued {
		t.Fatal("cursor move was not queued")
	}
	if row != 1 || column != 2 {
		t.Fatalf("got (%d,%d), want (1,2)", row, column)
	}
}

type paintComponent struct {
	tux.Atomic

	b rune
}

func (c paintComponent) Render(_ tux.BuildContext, ctx tux.RenderContext) error {
	ctx.Paint(0, 0, tux.Cell{Ch: c.b, Width: 1})
	return nil
}
