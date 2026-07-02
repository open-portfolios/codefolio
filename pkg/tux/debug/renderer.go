package debug

import "github.com/open-portfolios/codefolio/pkg/tux"

var (
	_ tux.RenderContext = (*Renderer)(nil)
)

type Renderer struct {
	Buf          [][]tux.Cell
	cursorQueued bool
	cursorRow    int
	cursorColumn int
}

func NewRenderer(row, column int) *Renderer {
	buf := make([][]tux.Cell, row)
	for i := range buf {
		buf[i] = make([]tux.Cell, column)
	}
	return &Renderer{Buf: buf}
}

func (r *Renderer) Render(ctx tux.BuildContext, root tux.Component) error {
	return root.Render(ctx, r)
}

func (r *Renderer) String(row, start, end int) string {
	line := make([]byte, end-start)
	for column := start; column < end; column++ {
		line[column-start] = r.Buf[row][column].Ch
	}
	return string(line)
}

func (r *Renderer) CursorPos() (row, column int, queued bool) {
	return r.cursorRow, r.cursorColumn, r.cursorQueued
}

func (r *Renderer) Paint(row, column int, cell tux.Cell) { r.Buf[row][column] = cell }

func (r *Renderer) QueueCursorMove(row, column int) {
	r.cursorQueued = true
	r.cursorRow = row
	r.cursorColumn = column
}

func (r *Renderer) Submit() error { return nil }
