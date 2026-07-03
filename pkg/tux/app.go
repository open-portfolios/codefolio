package tux

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/open-portfolios/codefolio/pkg/tux/platform"
)

// App is the top-level application driver. It owns the render loop, raw
// terminal mode, and input dispatch.
type App struct {
	root     Component
	rows     int
	columns  int
	renderer *Renderer

	// dirtyCh is a buffered channel of capacity 1 used as a non-blocking
	// dirty flag. Sending to it marks the app as needing a redraw; the
	// channel's capacity ensures concurrent senders never block.
	dirtyCh  chan struct{}
	stopCh   chan struct{}
	stopOnce sync.Once

	frameDuration time.Duration
	onKeyboard    func(*App, KeyboardEvent)
	onError       func(error)

	focused      Component // currently focused component
	focusApplied bool      // whether an AutoFocus component has been applied in this render pass

	buildEpoch uint64 // incremented before each Build pass
	keyboard   keyboardListener
	terminal   *platform.State
}

type AppOption func(*App)

func NewApp(root Component, options ...AppOption) *App {
	app := &App{
		root:          root,
		rows:          24,
		columns:       80,
		dirtyCh:       make(chan struct{}, 1),
		stopCh:        make(chan struct{}),
		frameDuration: time.Second / 30,
		onError: func(err error) {
			panic(fmt.Sprintf("tux: unhandled error: %v", err))
		},
	}
	for _, option := range options {
		option(app)
	}
	app.renderer = NewRenderer(app.rows, app.columns)
	return app
}

func WithSize(rows, columns int) AppOption {
	return func(app *App) {
		app.rows = rows
		app.columns = columns
	}
}

func WithFrameRate(fps int) AppOption {
	return func(app *App) {
		if fps <= 0 {
			return
		}
		app.frameDuration = time.Second / time.Duration(fps)
	}
}

func WithKeyboard(fn func(*App, KeyboardEvent)) AppOption {
	return func(app *App) {
		app.onKeyboard = fn
	}
}

func WithErrorHandler(fn func(error)) AppOption {
	return func(app *App) {
		app.onError = fn
	}
}

// Run opens the terminal, starts the render loop, and blocks until Stop is
// called or an error occurs.
func (a *App) Run() error {
	if err := a.Open(); err != nil {
		return err
	}
	defer a.Close()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	defer signal.Stop(signalCh)

	inputCh := make(chan byte, 32)
	go readInput(inputCh, a.stopCh)

	go func() {
		select {
		case <-signalCh:
			a.Stop()
		case <-a.stopCh:
		}
	}()

	a.MarkDirty()

	ticker := time.NewTicker(a.frameDuration)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return nil
		case <-ticker.C:
			// Drain all available input bytes before rendering so the frame
			// reflects the most recent user input.
		drain:
			for {
				select {
				case b := <-inputCh:
					for _, e := range a.keyboard.handleByte(b) {
						// Check for Ctrl+C
						if e.Mod == ModCtrl && e.Key == KeyRune && e.Rune == 'c' {
							a.Stop()
							break drain
						}
						a.dispatchKeyboard(e)
					}
				default:
					break drain
				}
			}

			// Flush any ESC sequence that has waited past its deadline.
			for _, e := range a.keyboard.flush() {
				a.dispatchKeyboard(e)
			}

			// Render once if the frame is dirty.
			select {
			case <-a.dirtyCh:
				if err := a.Render(); err != nil {
					return err
				}
			default:
			}
		}
	}
}

// Open switches the terminal to raw mode and enters the alternate screen.
func (a *App) Open() error {
	state, err := platform.EnterRawMode()
	if err != nil {
		return err
	}
	a.terminal = state
	_, err = os.Stdout.WriteString("\x1b[?1049h\x1b[2J\x1b[H")
	return err
}

// Close restores the terminal and leaves the alternate screen.
func (a *App) Close() error {
	_, writeErr := os.Stdout.WriteString("\x1b[?1049l")
	rawErr := platform.ExitRawMode(a.terminal)
	a.terminal = nil
	if writeErr != nil {
		return writeErr
	}
	return rawErr
}

// Stop signals the run loop to exit cleanly.
func (a *App) Stop() {
	a.stopOnce.Do(func() { close(a.stopCh) })
}

// MarkDirty schedules a redraw for the next frame. Safe to call from any
// goroutine; concurrent calls are coalesced — the channel has capacity 1.
func (a *App) MarkDirty() {
	select {
	case a.dirtyCh <- struct{}{}:
	default:
	}
}

// Render rebuilds the component tree and submits a diff to the terminal.
// It increments the build epoch so State.Get(ctx) subscriptions from previous
// passes are not duplicated.
func (a *App) Render() error {
	a.buildEpoch++
	a.focusApplied = false // Reset AutoFocus flag for this render pass
	ctx := BuildContext{
		notify: a.MarkDirty,
		epoch:  a.buildEpoch,
		app:    a,
	}

	// Walk the Build chain until we reach an atomic component (Build returns nil).
	root := a.root
	for {
		child := root.Build(ctx)
		if child == nil {
			break
		}
		if child == root {
			panic("tux: composition cycle detected")
		}
		root = child
	}

	if a.renderer == nil {
		return fmt.Errorf("tux: renderer is nil")
	}

	a.renderer.Clear()
	return a.renderer.Render(ctx, root)
}

// SetFocus sets the currently focused component.
func (a *App) SetFocus(component Component) {
	a.focused = component
	a.MarkDirty()
}

// GetFocus returns the currently focused component.
func (a *App) GetFocus() Component {
	return a.focused
}

// MarkFocusApplied marks that an AutoFocus component has been applied in this render pass.
// This is used internally to ensure only the first AutoFocus component gets focus.
func (a *App) MarkFocusApplied() {
	a.focusApplied = true
}

// IsFocusApplied returns whether an AutoFocus component has already been applied in this render pass.
func (a *App) IsFocusApplied() bool {
	return a.focusApplied
}

func (a *App) dispatchKeyboard(event KeyboardEvent) {
	// Prioritize focused component
	if a.focused != nil {
		if handler, ok := a.focused.(KeyboardHandler); ok {
			if err := handler.OnKeyboard(event); err != nil {
				if a.onError != nil {
					a.onError(err)
				}
			}
			return // Event consumed
		}
	}

	// Fallback to global keyboard callback
	if a.onKeyboard != nil {
		a.onKeyboard(a, event)
	}
}

func readInput(ch chan<- byte, stopCh <-chan struct{}) {
	buf := make([]byte, 1)
	for {
		select {
		case <-stopCh:
			return
		default:
		}

		n, err := os.Stdin.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			continue
		}
		if n == 0 {
			continue
		}

		select {
		case ch <- buf[0]:
		case <-stopCh:
			return
		}
	}
}
