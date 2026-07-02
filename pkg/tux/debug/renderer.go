package debug

import "github.com/open-portfolios/codefolio/pkg/tux"

var (
	_ tux.RenderContext = (*Renderer)(nil)
)

type Renderer struct {
	Buf [][]byte
}

func NewRenderer(row, column int) *Renderer {
	buf := make([][]byte, row)
	for i := range buf {
		buf[i] = make([]byte, column)
	}
	return &Renderer{Buf: buf}
}

func (r *Renderer) Render(ctx tux.BuildContext, root tux.Component) error {
	return root.Render(ctx, r)
}

func (r *Renderer) Paint(row, column int, b byte) { r.Buf[row][column] = b }
func (r *Renderer) Flush() error                  { return nil }
