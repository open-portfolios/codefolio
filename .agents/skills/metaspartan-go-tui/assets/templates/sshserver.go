// sshserver.go
//
// Multi-tenant TUI server: each SSH session gets an isolated *ui.Backend
// and its own rendered dashboard. Built on gliderlabs/ssh + tcell v3 Tty.
//
// Build:  go mod init example && go mod tidy && go run sshserver.go
// Test:   ssh -tt -p 2222 user@localhost   (then any password)
//
// IMPORTANT: Generate a host key first:
//   ssh-keygen -t ed25519 -f hostkey -N ""
//
// Quit inside the TUI: press q, <C-c>, or <Escape>.

package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gliderlabs/ssh"

	ui "github.com/metaspartan/gotui/v5"
	"github.com/metaspartan/gotui/v5/widgets"
)

// ---------------------------------------------------------------------------
// 1. TTY adapter: bridge gliderlabs/ssh session -> tcell v3 Tty interface
// ---------------------------------------------------------------------------

type sessionTTY struct {
	sess ssh.Session

	mu       sync.RWMutex
	w, h     int
	resizeCh chan<- bool

	winCh  <-chan ssh.Window
	closed chan struct{}
}

func newSessionTTY(sess ssh.Session) (*sessionTTY, error) {
	pty, winCh, ok := sess.Pty()
	if !ok {
		return nil, fmt.Errorf("no PTY requested (try: ssh -tt host -p 2222)")
	}

	t := &sessionTTY{
		sess:   sess,
		w:      pty.Window.Width,
		h:      pty.Window.Height,
		winCh:  winCh,
		closed: make(chan struct{}),
	}

	// Forward ssh window-change notifications into tcell's resize channel.
	go func() {
		for {
			select {
			case <-t.closed:
				return
			case win, ok := <-t.winCh:
				if !ok {
					return
				}
				t.mu.Lock()
				t.w, t.h = win.Width, win.Height
				ch := t.resizeCh
				t.mu.Unlock()
				if ch != nil {
					select {
					case ch <- true:
					default:
					}
				}
			}
		}
	}()

	return t, nil
}

// Size returns the current PTY dimensions.
func (t *sessionTTY) Size() (int, int) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.w, t.h
}

// NotifyResize registers tcell's resize channel.
func (t *sessionTTY) NotifyResize(ch chan<- bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.resizeCh = ch
}

// WindowSize is required by tcell v3's Tty interface.
func (t *sessionTTY) WindowSize() (int, int, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.w, t.h, nil
}

// Read / Write / Close delegate to the ssh.Session.
func (t *sessionTTY) Read(p []byte) (int, error)  { return t.sess.Read(p) }
func (t *sessionTTY) Write(p []byte) (int, error) { return t.sess.Write(p) }

func (t *sessionTTY) Close() error {
	select {
	case <-t.closed:
	default:
		close(t.closed)
	}
	return t.sess.Close()
}

// ---------------------------------------------------------------------------
// 2. Build the dashboard UI once per session (each session has its own Backend)
// ---------------------------------------------------------------------------

type dashboard struct {
	app  *ui.Backend
	grid *ui.Grid
}

func newDashboard() *dashboard {
	d := &dashboard{}

	// Header
	p := widgets.NewParagraph()
	p.Title = "Dashboard"
	p.Text = "SSH session | press q to disconnect"
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorLightCyan
	p.TitleStyle = ui.NewStyle(ui.ColorLightCyan, ui.ColorClear, ui.ModifierBold)
	p.TitleAlignment = ui.AlignCenter
	p.TitleRight = "ssh"
	p.BorderRounded = false

	// Logs list (the only thing we render for this minimal example)
	l := widgets.NewList()
	l.Title = "Events"
	l.Rows = []string{
		"[INFO] Connected",
		"[INFO] Press q to quit",
	}
	l.TextStyle.Fg = ui.ColorYellow
	l.SelectedStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow)
	l.TitleStyle.Fg = ui.ColorYellow
	l.BorderStyle.Fg = ui.ColorYellow
	l.TitleBottom = "live"
	l.TitleBottomAlignment = ui.AlignRight
	l.BorderRounded = true

	d.grid = ui.NewGrid()
	d.grid.Set(
		ui.NewRow(1.0/8, ui.NewCol(1.0, p)),
		ui.NewRow(7.0/8, ui.NewCol(1.0, l)),
	)
	return d
}

func (d *dashboard) onResize(w, h int) {
	d.grid.SetRect(0, 0, w, h)
	d.app.Clear()
	d.app.Render(d.grid)
}

// ---------------------------------------------------------------------------
// 3. Per-session runner
// ---------------------------------------------------------------------------

func runSession(sess ssh.Session) {
	tty, err := newSessionTTY(sess)
	if err != nil {
		fmt.Fprintln(sess.Stderr(), err)
		return
	}
	defer tty.Close()

	// CRITICAL: ui.NewBackend (not the package-level helpers) gives each
	// session its own isolated tcell.Screen. Never use ui.Init() in
	// per-session code paths.
	app, err := ui.NewBackend(&ui.InitConfig{CustomTTY: tty})
	if err != nil {
		fmt.Fprintln(sess.Stderr(), "init:", err)
		return
	}
	defer app.Close()

	d := newDashboard()
	d.app = app

	w, h := tty.Size()
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	d.onResize(w, h)

	events := app.PollEventsWithContext(sess.Context())
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sess.Context().Done():
			return

		case e, ok := <-events:
			if !ok {
				return
			}
			switch e.ID {
			case "q", "<C-c>", "<Escape>":
				return
			case "<Resize>":
				r := e.Payload.(ui.Resize)
				d.onResize(r.Width, r.Height)
			}

		case <-ticker.C:
			d.app.Render(d.grid)
		}
	}
}

// ---------------------------------------------------------------------------
// 4. Main: SSH server entry point
// ---------------------------------------------------------------------------

func main() {
	ssh.Handle(runSession)
	log.Fatal(ssh.ListenAndServe(":2222", nil,
		ssh.HostKeyFile("hostkey"),
		ssh.PasswordAuth(func(ctx ssh.Context, pass string) bool {
			// TODO: replace with a real auth check.
			return pass == "letmein"
		}),
	))
}
