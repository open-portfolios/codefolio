# SSH Server Pattern

gotui v5 was designed to be drop-in for **multi-tenant SSH TUI servers**:
each SSH session gets its own isolated `*ui.Backend` (and thus its own
`tcell.Screen`) without conflicting with other sessions.

## The Critical Rule

> **Never call `ui.Init()` in per-session code.** `ui.Init` wires up a
> global `DefaultBackend`. For SSH servers, always use `ui.NewBackend`
> with a per-session TTY adapter.

```go
// Per-session:
app, err := ui.NewBackend(&ui.InitConfig{CustomTTY: tty})
if err != nil { return err }
defer app.Close()
```

## Components

### 1. TTY adapter (the tricky part)

tcell v3 expects a `Tty` interface:

```go
type Tty interface {
    io.ReadWriter                // Read(p) / Write(p)
    Size() (int, int)
    WindowSize() (int, int, error) // v3 explicit
    NotifyResize(ch chan<- bool)
    Close() error
}
```

The `ssh.Session` from `gliderlabs/ssh` gives you `Read`, `Write`, `Pty()`
returning a `Window`, and a `WindowChanges` channel. You wrap it:

```go
type sessionTTY struct {
    sess     ssh.Session
    winCh    <-chan ssh.Window
    resizeCh chan<- bool
    mu       sync.RWMutex
    w, h     int
    closed   chan struct{}
}

func (t *sessionTTY) Read(p []byte)  (int, error) { return t.sess.Read(p) }
func (t *sessionTTY) Write(p []byte) (int, error) { return t.sess.Write(p) }

func (t *sessionTTY) Size() (int, int)            { /* lock, return t.w, t.h */ }
func (t *sessionTTY) WindowSize() (int, int, error) { /* lock, return t.w, t.h, nil */ }

func (t *sessionTTY) NotifyResize(ch chan<- bool) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.resizeCh = ch
}

func (t *sessionTTY) Close() error { /* signal goroutine, close session */ }
```

The bridge goroutine:

```go
go func() {
    for win := range t.winCh {           // ssh sends new sizes
        t.mu.Lock()
        t.w, t.h = win.Width, win.Height
        ch := t.resizeCh                 // tcell's resize channel
        t.mu.Unlock()
        if ch != nil {
            select {
            case ch <- true:             // non-blocking signal
            default:
            }
        }
    }
}()
```

### 2. Per-session loop

Don't use `app.Run()` for SSH — the context is per-session:

```go
func runSession(sess ssh.Session) {
    tty, err := newSessionTTY(sess)
    if err != nil {
        fmt.Fprintln(sess.Stderr(), err)
        return
    }
    defer tty.Close()

    app, err := ui.NewBackend(&ui.InitConfig{CustomTTY: tty})
    if err != nil { return }
    defer app.Close()

    grid := buildGrid()
    w, h := tty.Size()
    grid.SetRect(0, 0, w, h)
    app.Render(grid)

    events := app.PollEventsWithContext(sess.Context())
    ticker := time.NewTicker(250 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-sess.Context().Done():
            return                  // SSH disconnect
        case e := <-events:
            switch e.ID {
            case "q", "<C-c>", "<Escape>":
                return
            case "<Resize>":
                r := e.Payload.(ui.Resize)
                grid.SetRect(0, 0, r.Width, r.Height)
                app.Clear()
                app.Render(grid)
            }
        case <-ticker.C:
            app.Render(grid)
        }
    }
}
```

**Note**: `PollEventsWithContext(ctx)` is the SSH-friendly variant — it
returns when the context is cancelled. Use `app.PollEvents()` only when
you don't need context cancellation.

### 3. Main: glue gliderlabs/ssh

```go
import "github.com/gliderlabs/ssh"

func main() {
    ssh.Handle(runSession)
    log.Fatal(ssh.ListenAndServe(":2222", nil,
        ssh.HostKeyFile("hostkey"),
        ssh.PasswordAuth(func(ctx ssh.Context, pass string) bool {
            return pass == "letmein"
        }),
    ))
}
```

**Generate a host key first**:
```bash
ssh-keygen -t ed25519 -f hostkey -N ""
```

**Test**:
```bash
ssh -tt -p 2222 user@localhost
# (password: letmein)
```

## Render Cadence Over SSH

SSH adds network latency on every frame. Recommended rates:

| Content type          | Tick interval     |
| --------------------- | ----------------- |
| Static / rare updates | event-driven only |
| Status (CPU / memory) | 500–1000 ms       |
| Animated charts       | 150–250 ms        |
| Real-time (typing)    | event-driven      |

Going faster than 60 ms will flicker on slow links.

## Common Pitfalls

### Wrong: shared backend
```go
ui.Init()           // global DefaultBackend — only one allowed
ui.NewBackend(...)  // this fails or conflicts
```

### Right: per-session backend
```go
app, _ := ui.NewBackend(&ui.InitConfig{CustomTTY: tty})
defer app.Close()
```

### Wrong: blocking event pump
```go
for e := range app.PollEvents() { ... }  // ignores sess.Context()
```
On disconnect, `PollEvents` may not return immediately.

### Right: context-aware pump
```go
events := app.PollEventsWithContext(sess.Context())
for {
    select {
    case <-sess.Context().Done(): return
    case e := <-events: ...
    }
}
```

### Wrong: per-session state in package globals
```go
var globalGrid *ui.Grid  // shared across sessions
```

### Right: per-session state in closure
```go
func runSession(sess ssh.Session) {
    grid := buildGrid()  // one grid per session
    ...
}
```

## Authentication

`ssh.PasswordAuth` is the simplest. For keys:

```go
ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
    // check key against allow-list
    return true
})
```

For production: front the server with a proper authenticator and audit
log.

## Multiple Sessions & Performance

- Each session is independent goroutine + state. No locking needed across
  sessions.
- Allocations per session: ~1 Grid, several widgets, ~one buffer. Cheap.
- Use a moderate render cadence (150–250 ms for animated dashboards) to avoid
  excessive network output.

## See Also

- `assets/templates/sshserver.go` — runnable starting point.
- `references/api-styles.md` — explains why SSH uses Backend-only style.