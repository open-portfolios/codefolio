package tux

import (
	"bufio"
	"fmt"
	"os"
)

var (
	_ RenderContext = (*Renderer)(nil)
)

type Renderer struct {
	buf          [][]Cell
	cursorQueued bool
	cursorRow    int
	cursorColumn int
}

func NewRenderer(row, column int) *Renderer {
	buf := make([][]Cell, row)
	for i := range buf {
		buf[i] = make([]Cell, column)
	}
	return &Renderer{buf: buf}
}

func (r *Renderer) Render(ctx BuildContext, root Component) error {
	if err := root.Render(ctx, r); err != nil {
		return err
	}
	return r.Submit()
}

func (r *Renderer) Paint(row, column int, cell Cell) { r.buf[row][column] = cell }

func (r *Renderer) QueueCursorMove(row, column int) {
	r.cursorQueued = true
	r.cursorRow = row
	r.cursorColumn = column
}

func (r *Renderer) Submit() error {
	w := bufio.NewWriter(os.Stdout)
	for row, cells := range r.buf {
		if row > 0 {
			if err := w.WriteByte('\n'); err != nil {
				return err
			}
		}

		for _, cell := range cells {
			ch := cell.Ch
			if ch == 0 {
				ch = ' '
			}
			if err := w.WriteByte(ch); err != nil {
				return err
			}
		}
	}

	if r.cursorQueued {
		if _, err := fmt.Fprintf(w, "\x1b[%d;%dH", r.cursorRow+1, r.cursorColumn+1); err != nil {
			return err
		}
	}

	return w.Flush()
}
