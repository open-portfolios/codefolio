package tux

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

type App struct {
	root     Component
	rows     int
	columns  int
	renderer *Renderer

	dirty         atomic.Bool
	stopCh        chan struct{}
	stopOnce      sync.Once
	frameDuration time.Duration
	onInput       func(*App, InputEvent)
	terminal      *terminalState
	escPending    bool
	escBuffer     []byte
	escDeadline   time.Time
	extPending    bool
	utf8Buffer    []byte
}

type AppOption func(*App)

func NewApp(root Component, options ...AppOption) *App {
	app := &App{
		root:          root,
		rows:          24,
		columns:       80,
		stopCh:        make(chan struct{}),
		frameDuration: time.Second / 30,
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

func WithInput(fn func(*App, InputEvent)) AppOption {
	return func(app *App) {
		app.onInput = fn
	}
}

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

	for {
		frameStart := time.Now()

	drain:
		for {
			select {
			case b := <-inputCh:
				if b == 3 {
					a.Stop()
					break drain
				}
				a.handleInputByte(b)
			default:
				break drain
			}
		}
		a.flushPendingEsc()

		select {
		case <-a.stopCh:
			return nil
		default:
			if err := a.Render(); err != nil {
				return err
			}
		}

		remaining := a.frameDuration - time.Since(frameStart)
		if remaining < 0 {
			remaining = 0
		}

		select {
		case <-a.stopCh:
			return nil
		case <-time.After(remaining):
		}
	}
}

func (a *App) Open() error {
	state, err := enterRawMode()
	if err != nil {
		return err
	}
	a.terminal = state
	_, err = os.Stdout.WriteString("\x1b[?1049h\x1b[2J\x1b[H")
	return err
}

func (a *App) Close() error {
	_, writeErr := os.Stdout.WriteString("\x1b[?1049l")
	rawErr := exitRawMode(a.terminal)
	a.terminal = nil
	if writeErr != nil {
		return writeErr
	}
	return rawErr
}

func (a *App) Stop() {
	a.stopOnce.Do(func() {
		close(a.stopCh)
	})
}

func (a *App) MarkDirty() {
	a.dirty.Store(true)
}

func (a *App) Render() error {
	if !a.checkAndClearDirty() {
		return nil
	}

	ctx := BuildContext{}
	root := a.root
	for {
		artifact := root.Build(ctx)
		if artifact == nil {
			break
		}
		if artifact == root {
			panic("composition cycle is not allowed")
		}
		root = artifact
	}

	if a.renderer == nil {
		return fmt.Errorf("tux: app renderer is nil")
	}

	a.renderer.Clear()
	return a.renderer.Render(ctx, root)
}

func (a *App) checkAndClearDirty() bool {
	return a.dirty.Swap(false)
}

func (a *App) handleInputByte(b byte) {
	if a.extPending {
		a.extPending = false
		switch b {
		case 'H':
			a.dispatchInput(InputEvent{Key: KeyUp})
		case 'P':
			a.dispatchInput(InputEvent{Key: KeyDown})
		case 'K':
			a.dispatchInput(InputEvent{Key: KeyLeft})
		case 'M':
			a.dispatchInput(InputEvent{Key: KeyRight})
		}
		return
	}

	if a.escPending {
		a.escBuffer = append(a.escBuffer, b)
		if len(a.escBuffer) == 1 {
			if b != '[' {
				a.escPending = false
				a.escBuffer = nil
			}
			return
		}

		if len(a.escBuffer) == 2 {
			if a.escBuffer[1] >= '0' && a.escBuffer[1] <= '9' {
				return
			}
			switch a.escBuffer[1] {
			case 'A':
				a.dispatchInput(InputEvent{Key: KeyUp})
			case 'B':
				a.dispatchInput(InputEvent{Key: KeyDown})
			case 'C':
				a.dispatchInput(InputEvent{Key: KeyRight})
			case 'D':
				a.dispatchInput(InputEvent{Key: KeyLeft})
			}
			a.escPending = false
			a.escBuffer = nil
		}

		if len(a.escBuffer) == 3 {
			if a.escBuffer[1] == '3' && a.escBuffer[2] == '~' {
				a.dispatchInput(InputEvent{Key: KeyDelete})
			}
			a.escPending = false
			a.escBuffer = nil
		}
		return
	}

	switch b {
	case 0, 224:
		a.extPending = true
	case 27:
		a.escPending = true
		a.escBuffer = nil
		a.escDeadline = time.Now().Add(30 * time.Millisecond)
	case 8, 127:
		a.dispatchInput(InputEvent{Key: KeyBackspace, Rune: rune(b)})
	case '\r', '\n':
		a.dispatchInput(InputEvent{Key: KeyEnter, Rune: rune(b)})
	default:
		a.handleRuneByte(b)
	}
}

func (a *App) handleRuneByte(b byte) {
	if b < utf8.RuneSelf && len(a.utf8Buffer) == 0 {
		a.dispatchInput(InputEvent{Key: KeyRune, Rune: rune(b)})
		return
	}

	a.utf8Buffer = append(a.utf8Buffer, b)
	if !utf8.FullRune(a.utf8Buffer) {
		return
	}
	r, size := utf8.DecodeRune(a.utf8Buffer)
	if r != utf8.RuneError || size > 1 {
		a.dispatchInput(InputEvent{Key: KeyRune, Rune: r})
	}
	a.utf8Buffer = nil
}

func (a *App) flushPendingEsc() {
	if !a.escPending || time.Now().Before(a.escDeadline) {
		return
	}
	a.escPending = false
	a.escBuffer = nil
}

func (a *App) dispatchInput(event InputEvent) {
	if a.onInput != nil {
		a.onInput(a, event)
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
