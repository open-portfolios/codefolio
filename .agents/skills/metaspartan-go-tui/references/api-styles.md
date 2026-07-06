# API Styles: Application (OO) vs Functional

gotui v5 deliberately ships **two API styles** that share the same backend.
Pick the one that matches your use case — both can coexist in the same
program if you really need to.

## At a Glance

| Concern        | Functional API                           | Application (OO) API                                    |
| -------------- | ---------------------------------------- | ------------------------------------------------------- |
| Init           | `ui.Init()`                              | `ui.NewApp()`                                           |
| Render         | `ui.Render(items...)`                    | `app.Backend.Render(items...)`                          |
| Close          | `ui.Close()`                             | `app.Run()` (deferred internally)                       |
| Events         | `for e := range ui.PollEvents() { ... }` | `app.Run()` blocks; you handle events via `HandleEvent` |
| Resize         | Manual: `case "<Resize>": ...`           | Automatic                                               |
| Quit keys      | Manual                                   | Automatic (`q`, `<C-c>`)                                |
| Focus          | n/a                                      | `app.SetFocus(widget)`                                  |
| Multi-instance | **No** (global state)                    | **Yes** via `NewBackend(CustomTTY: ...)`                |

## Functional API — when to use

- Quick demos, scripts, monitoring tools.
- Programs that own a single screen and a single event loop.
- Code that needs to be drop-in compatible with classic `termui`.

```go
if err := ui.Init(); err != nil { log.Fatal(err) }
defer ui.Close()

grid := buildGrid()
w, h := ui.TerminalDimensions()
grid.SetRect(0, 0, w, h)
ui.Render(grid)

for e := range ui.PollEvents() {
    switch e.ID {
    case "q", "<C-c>":
        return
    case "<Resize>":
        r := e.Payload.(ui.Resize)
        grid.SetRect(0, 0, r.Width, r.Height)
        ui.Clear()
        ui.Render(grid)
    }
}
```

**Caveat**: `ui.Init`/`ui.Close`/`ui.Render`/`ui.PollEvents` all touch a
**global** `DefaultBackend`. You cannot run two of these in the same
process at the same time. SSH servers must use the OO API instead.

## Application (OO) API — when to use

- Anything new. This is the **recommended** entry point.
- Multi-widget apps where widgets consume keyboard focus.
- Apps that want automatic resize + quit handling.
- SSH / multi-tenant servers (because `app.Backend` is per-instance).

```go
type MyWidget struct {
    *widgets.Paragraph
}

func (w *MyWidget) HandleEvent(e ui.Event) bool {
    switch e.ID {
    case "<Space>":
        w.Text = "Pressed!"
        return true
    }
    return false
}

func main() {
    app := ui.NewApp()
    app.SetRoot(&MyWidget{Paragraph: widgets.NewParagraph()}, true)
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### What `Run()` gives you for free

1. Initial `SetRect(0, 0, w, h)` from current terminal size.
2. Initial `Render(root)`.
3. Event loop dispatching to `focus.HandleEvent` → `root.HandleEvent`.
4. Default quit on `q` and `<C-c>`.
5. Resize: re-`SetRect` + `Clear` + `Render`.

### When you want to drive the loop yourself

Use `ui.NewBackend(&ui.InitConfig{...})` directly — same backend that
`Application` uses internally, but you write the loop:

```go
app, err := ui.NewBackend(&ui.InitConfig{})
if err != nil { return err }
defer app.Close()

uiEvents := app.PollEvents()
ticker := time.NewTicker(100 * time.Millisecond)
defer ticker.Stop()

for {
    select {
    case e := <-uiEvents:
        // your switch
    case <-ticker.C:
        app.Clear()
        app.Render(root)
    }
}
```

This is the form to use for **SSH servers** (see
`assets/templates/sshserver.go`).

## Decision Tree

```
Need multiple TUIs in one process (SSH / tests)?
  └─ yes → OO API, call ui.NewBackend per instance
  └─ no  → either works; default to OO

Need automatic resize + quit?
  └─ yes → Application.Run()
  └─ no  → either

Widget needs keyboard focus?
  └─ yes → OO API (functional has no focus model)
  └─ no  → either

Migrating from termui?
  └─ yes → start with functional API, gradually move to OO
  └─ no  → start with OO
```

## Mixing Styles (rare)

You can call `ui.Render(x)` from inside an Application — they share the
same `DefaultBackend`. But once you start calling `ui.Init()`,
`DefaultBackend` is locked to that mode. For SSH scenarios, **always**
use `ui.NewBackend` per session and never touch `ui.Init`.