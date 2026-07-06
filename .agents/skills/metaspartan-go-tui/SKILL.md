---
name: metaspartan-go-tui
description: >
  Scaffold Go terminal applications using metaspartan/gotui v5 (tcell v3-based
  immediate-mode TUI library). Use when building Go terminal dashboards,
  monitoring tools, CLI data viewers, SSH-hosted TUIs, or any program requiring
  rich terminal widgets (paragraph, list, table, tabs, spinner, charts). Covers
  the Application OO API, Drawable/Block/Buffer architecture, Grid/Flex layout,
  event handling, theming, gradient styling, headless PNG screenshot mode, and
  multi-tenant SSH server patterns. Do NOT use for bubbletea/lipgloss-based
  projects (different paradigm: Elm-architecture vs immediate-mode), nor for
  pure tmux-style TUIs, nor for non-Go projects.
---

# metaspartan-go-tui

Scaffold Go TUIs with **`github.com/metaspartan/gotui/v5`** — a tcell v3-based
immediate-mode library with a rich widget set, true-color support, and first-class
SSH server support.

This skill is self-contained: use the instructions, local reference pages, and
templates in this directory without consulting anything else.

## When to Use This Skill

| If you want to build…                                         | Use this skill?         |
| ------------------------------------------------------------- | ----------------------- |
| A Go terminal dashboard (system monitor, log viewer, data UI) | **Yes**                 |
| A Go CLI tool with rich interactive UI                        | **Yes**                 |
| A multi-tenant SSH TUI server in Go                           | **Yes**                 |
| An Elm-architecture Go TUI (Model/Update/View + messages)     | **No** — use bubbletea  |
| A Rust / Python / Node TUI                                    | **No** — wrong language |
| A simple stdin/stdout CLI with no UI                          | **No** — overkill       |

## Library Identity

|         |                                       |
| ------- | ------------------------------------- |
| Module  | `github.com/metaspartan/gotui/v5`     |
| Go      | ≥ 1.24                                |
| License | MIT                                   |
| Backend | `github.com/gdamore/tcell/v3`         |
| Style   | Immediate-mode (not Elm-architecture) |
| Widgets | 27 in `widgets/` subpackage           |

## 5-Minute Hello World

```go
package main

import (
    "log"
    ui "github.com/metaspartan/gotui/v5"
    "github.com/metaspartan/gotui/v5/widgets"
)

type Hello struct{ *widgets.Paragraph }

func (h *Hello) HandleEvent(e ui.Event) bool {
    if e.Type == ui.KeyboardEvent && e.ID == "<Space>" {
        h.Text = "Pressed!"
        return true
    }
    return false
}

func main() {
    app := ui.NewApp()
    app.SetRoot(&Hello{widgets.NewParagraph()}, true)
    log.Fatal(app.Run())
}
```

Run: `go run .`  Quit: `q` or `<C-c>`.

Full scaffold: `assets/templates/helloworld.go`.

## API Style: Pick One

| Style                              | Entry                                           | Best for                              |
| ---------------------------------- | ----------------------------------------------- | ------------------------------------- |
| **Application (OO)** ← recommended | `ui.NewApp()` / `app.Run()`                     | New apps, focus handling, auto-resize |
| Functional                         | `ui.Init()` / `ui.Render()` / `ui.PollEvents()` | Demos, termui migrations              |
| Backend-only (manual loop)         | `ui.NewBackend(&ui.InitConfig{...})`            | SSH servers, tests                    |

**Rule of thumb**: start with OO. Switch to Backend-only when you need
multi-instance support (e.g. SSH servers).

→ See `references/api-styles.md` for the full decision tree.

## Architecture: Three Building Blocks

1. **`Block`** — every widget embeds this. Provides borders, titles, padding,
   `Inner` rectangle math. Custom widgets start here.
2. **`Buffer` + `Cell{rune, style}`** — the rendering surface. `Render()` builds
   a union-sized buffer, asks each drawable to fill it, then flushes to tcell
   in one pass. No diffing, no flicker.
3. **`Application` / `Backend`** — the host. Application wraps a single
   Backend + root widget + focus widget + event loop. `NewBackend` is the
   per-instance variant used for SSH.

→ See `references/architecture.md`.

## Layout: Three Strategies

| Strategy   | API                                                         | Use when                       |
| ---------- | ----------------------------------------------------------- | ------------------------------ |
| Absolute   | `w.SetRect(x1, y1, x2, y2)`                                 | Single widget or modal overlay |
| Ratio grid | `ui.NewGrid()` + `ui.NewRow(r, ...)` + `ui.NewCol(r, ...)`  | Multi-panel dashboards         |
| Flex       | `widgets.NewFlex()` + `flex.AddItem(w, fixed, prop, focus)` | Sidebar + main + statusbar     |

For 90% of dashboards, Grid is the right answer. Ratio `1.0/8 + 3.0/8 + ...`
sums to `1.0` for clarity (auto-normalized otherwise).

→ See `references/layout.md`.

## Events

gotui translates raw tcell keys into stable string IDs so your widget code
never depends on `tcell`:

| Input         | ID                                                 |
| ------------- | -------------------------------------------------- |
| `q`, `<C-c>`  | reserved (quit, OO API only)                       |
| Arrow keys    | `<Up>`, `<Down>`, `<Left>`, `<Right>`              |
| Ctrl-letter   | `<C-a>` through `<C-z>`                            |
| Alt-letter    | `<M-x>`                                            |
| Function keys | `<F1>` through `<F12>`                             |
| Mouse click   | `<MouseLeft>`, `<MouseMiddle>`, `<MouseRight>`     |
| Mouse wheel   | `<MouseWheelUp>`, `<MouseWheelDown>`               |
| Resize        | `<Resize>` with payload `ui.Resize{Width, Height}` |

```go
func (w *MyWidget) HandleEvent(e ui.Event) bool {
    switch e.Type {
    case ui.KeyboardEvent:
        switch e.ID {
        case "j", "<Down>": w.cursor++; return true
        case "k", "<Up>":   w.cursor--; return true
        }
    case ui.MouseEvent:
        m := e.Payload.(ui.Mouse)
        if e.ID == "<MouseLeft>" { w.activate(m.X, m.Y); return true }
    }
    return false
}
```

→ See `references/events.md` for the full ID table.

## Styling

Flat `ui.Style{Fg, Bg, Modifier}` — no chain like lipgloss.

```go
ui.NewStyle(ui.ColorCyan, ui.ColorBlack, ui.ModifierBold)
ui.NewRGBColor(255, 100, 50)
ui.HexToColor("#ff6432")
ui.GenerateGradient(ui.ColorRed, ui.ColorBlue, 16)
```

Inline DSL inside `Paragraph.Text`:
```
Hello [world](fg:red,mod:bold)!
```

Global theme singleton: `ui.Theme.Paragraph.Text.Fg = ui.ColorWhite` — simple
but stateful.

→ See `references/styling.md`.

## Widget Catalog (27 widgets)

| Category  | Widgets                                                                                                                         |
| --------- | ------------------------------------------------------------------------------------------------------------------------------- |
| Display   | `Paragraph`, `List`, `Table`, `Tabs`, `Scrollbar`                                                                               |
| Input     | `Input`, `TextArea`, `Button`, `Checkbox`                                                                                       |
| Progress  | `Gauge`, `LineGauge`, `Sparkline`, `SparklineGroup`                                                                             |
| Charts    | `Plot`, `BarChart`, `StackedBarChart`, `StepChart`, `PieChart`, `DonutChart`, `FunnelChart`, `RadarChart`, `Heatmap`, `Treemap` |
| Structure | `Tree`, `Calendar`                                                                                                              |
| Visual    | `Spinner`, `Image`, `Canvas` (braille), `Modal`                                                                                 |

→ See `references/widgets-catalog.md` for full reference.

**Note**: avoid the bundled brand-specific logo widget in product templates.
Use your own ASCII art or a `Paragraph` with a custom title instead.

## Common Patterns

### Animated dashboard

```go
go func() {
    ticker := time.NewTicker(100 * time.Millisecond)
    for range ticker.C {
        sparkline.Data = append(sparkline.Data[1:], newSample())
        ui.Render(grid)
    }
}()
```

### Headless PNG (for docs)

```go
if len(os.Args) > 1 && os.Args[1] == "-screenshot" {
    grid.SetRect(0, 0, 120, 40)
    ui.SaveImage("screenshot.png", 120, 40, grid)
    return
}
```

→ See `references/screenshot-mode.md`.

### SSH server (multi-tenant)

```go
ssh.Handle(func(sess ssh.Session) {
    tty, _ := newSessionTTY(sess)
    app, _ := ui.NewBackend(&ui.InitConfig{CustomTTY: tty})
    // ... render per-session dashboard ...
})
ssh.ListenAndServe(":2222", nil, ssh.HostKeyFile("hostkey"))
```

→ See `assets/templates/sshserver.go` and `references/ssh-server.md`.

## Files in This Skill

```
techstack/metaspartan-go-tui/
├── SKILL.md                      ← you are here
├── references/
│   ├── architecture.md           Drawable/Block/Buffer/Application internals
│   ├── api-styles.md             Application OO vs Functional vs Backend-only
│   ├── events.md                 Full canonical key ID table
│   ├── styling.md                Theme, Style, inline DSL, gradients
│   ├── layout.md                 SetRect / Grid / Flex decision tree
│   ├── widgets-catalog.md        All 27 widgets with usage snippets
│   ├── ssh-server.md             Per-session Backend + Tty adapter pattern
│   └── screenshot-mode.md        -screenshot flag + SimulationScreen
└── assets/
    ├── spinner-frames.md         17 ready-made spinner frame sets
    └── templates/
        ├── helloworld.go         Minimum runnable (Application OO)
        ├── dashboard.go          Multi-widget Grid + ticker (Functional)
        └── sshserver.go          Per-session SSH TUI (Backend-only)
```

## Quick Recipe

For most agents, the workflow is:

1. **Identify the UI**: list of widgets, layout, event model.
2. **Pick the API style** using the table above.
3. **Copy a template** from `assets/templates/` into the user's Go application
   as a starting point, then adapt it to that app's package layout, module path,
   state model, and commands. Do not run templates in-place from the skill
   directory.
4. **Add widgets** via `references/widgets-catalog.md`.
5. **Wire events** via the ID table in `references/events.md`.
6. **Style** with the patterns in `references/styling.md`.
7. **(Optional)** Add SSH support via `assets/templates/sshserver.go`.