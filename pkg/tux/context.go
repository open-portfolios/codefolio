package tux

import "github.com/open-portfolios/codefolio/pkg/tux/misc"

// BuildContext is passed through the component tree during the Build phase.
//
// notify is the callback registered by State.Get to signal that a rebuild is
// needed when state changes. epoch is a monotonically increasing counter
// incremented by App at the start of each Build pass; State uses it to avoid
// registering duplicate watchers within the same pass. app is a reference to
// the App instance, allowing components to access focus management and other
// app-level functionality.
type BuildContext struct {
	notify func()
	epoch  uint64
	app    *App
}

// App returns the App instance for this build context.
// Returns nil if the context was not created by an App.
func (ctx BuildContext) App() *App {
	return ctx.app
}

type Cell struct {
	Ch         rune
	Width      uint8
	Foreground misc.Color
	Background misc.Color
}

type RenderContext interface {
	Paint(row, column int, cell Cell)
	QueueCursorMove(row, column int)
	Submit() error
}
