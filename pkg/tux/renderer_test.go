package tux

import (
	"testing"
)

func TestRendererPrintsToStdout(t *testing.T) {
	renderer := NewRenderer(3, 16)

	renderer.Paint(0, 0, testCell('T'))
	renderer.Paint(0, 1, testCell('U'))
	renderer.Paint(0, 2, testCell('X'))
	renderer.Paint(1, 0, testCell('r'))
	renderer.Paint(1, 1, testCell('e'))
	renderer.Paint(1, 2, testCell('n'))
	renderer.Paint(1, 3, testCell('d'))
	renderer.Paint(1, 4, testCell('e'))
	renderer.Paint(1, 5, testCell('r'))
	renderer.Paint(2, 0, testCell('o'))
	renderer.Paint(2, 1, testCell('k'))

	if err := renderer.Submit(); err != nil {
		t.Fatal(err)
	}
}

func testCell(ch rune) Cell {
	return Cell{Ch: ch, Width: 1}
}
