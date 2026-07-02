package tux

type BuildContext struct{}

type RenderContext interface {
	Paint(row, column int, b byte)
	Flush() error
}
