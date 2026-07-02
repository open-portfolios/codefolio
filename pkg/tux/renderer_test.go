package tux

import (
	"testing"
)

func TestRendererPrintsToStdout(t *testing.T) {
	renderer := NewRenderer(3, 16)

	renderer.Paint(0, 0, NewCell('T', ColorDefault, ColorDefault))
	renderer.Paint(0, 1, NewCell('U', ColorDefault, ColorDefault))
	renderer.Paint(0, 2, NewCell('X', ColorDefault, ColorDefault))
	renderer.Paint(1, 0, NewCell('r', ColorDefault, ColorDefault))
	renderer.Paint(1, 1, NewCell('e', ColorDefault, ColorDefault))
	renderer.Paint(1, 2, NewCell('n', ColorDefault, ColorDefault))
	renderer.Paint(1, 3, NewCell('d', ColorDefault, ColorDefault))
	renderer.Paint(1, 4, NewCell('e', ColorDefault, ColorDefault))
	renderer.Paint(1, 5, NewCell('r', ColorDefault, ColorDefault))
	renderer.Paint(2, 0, NewCell('o', ColorDefault, ColorDefault))
	renderer.Paint(2, 1, NewCell('k', ColorDefault, ColorDefault))

	if err := renderer.Submit(); err != nil {
		t.Fatal(err)
	}
}
