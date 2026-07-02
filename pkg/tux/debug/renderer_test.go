package debug

import (
	"testing"

	"github.com/open-portfolios/codefolio/pkg/tux"
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

func TestRendererPaintWritesByte(t *testing.T) {
	renderer := NewRenderer(2, 3)

	renderer.Paint(1, 2, 'x')

	if got, want := renderer.Buf[1][2], byte('x'); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRendererFlush(t *testing.T) {
	renderer := NewRenderer(1, 1)

	if err := renderer.Flush(); err != nil {
		t.Fatal(err)
	}
}

func TestRendererRender(t *testing.T) {
	renderer := NewRenderer(1, 1)

	if err := renderer.Render(tux.BuildContext{}, paintComponent{b: 'x'}); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.Buf[0][0], byte('x'); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

type paintComponent struct {
	tux.Atomic

	b byte
}

func (c paintComponent) Render(_ tux.BuildContext, ctx tux.RenderContext) error {
	ctx.Paint(0, 0, c.b)
	return nil
}
