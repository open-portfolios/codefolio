package tux

import (
	"testing"
)

func TestRendererPrintsToStdout(t *testing.T) {
	renderer := NewRenderer(3, 16)

	renderer.Paint(0, 0, Cell{Ch: 'T'})
	renderer.Paint(0, 1, Cell{Ch: 'U'})
	renderer.Paint(0, 2, Cell{Ch: 'X'})
	renderer.Paint(1, 0, Cell{Ch: 'r'})
	renderer.Paint(1, 1, Cell{Ch: 'e'})
	renderer.Paint(1, 2, Cell{Ch: 'n'})
	renderer.Paint(1, 3, Cell{Ch: 'd'})
	renderer.Paint(1, 4, Cell{Ch: 'e'})
	renderer.Paint(1, 5, Cell{Ch: 'r'})
	renderer.Paint(2, 0, Cell{Ch: 'o'})
	renderer.Paint(2, 1, Cell{Ch: 'k'})

	if err := renderer.Flush(); err != nil {
		t.Fatal(err)
	}
}
