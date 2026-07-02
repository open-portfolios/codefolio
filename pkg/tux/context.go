package tux

type BuildContext struct{}

type Cell struct {
	Ch         rune
	Width      uint8
	Foreground Color
	Background Color
}

type RenderContext interface {
	Paint(row, column int, cell Cell)
	QueueCursorMove(row, column int)
	Submit() error
}
