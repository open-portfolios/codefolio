package tux

type BuildContext struct{}

type Cell struct {
	Ch         byte
	Foreground Color
	Background Color
}

type RenderContext interface {
	Paint(row, column int, cell Cell)
	QueueCursorMove(row, column int)
	Submit() error
}
