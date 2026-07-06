# Architecture

## Mental Model

gotui is an **immediate-mode** TUI library built on **tcell v3**. It does not
use an Elm-style Model-Update-View loop. Instead, every frame you call
`ui.Render(items...)` (or `app.Backend.Render(items...)`) and the library:

1. Computes the union bounding box of all items.
2. Asks each item to fill a shared `Buffer` of `Cell{rune, style}`.
3. Pushes the buffer to the terminal in a single pass.

There is **no automatic diffing** and **no retained tree**. You decide what
to draw each frame.

## Core Types

```go
// Every widget implements Drawable.
type Drawable interface {
    image.Rectangle       // x1, y1, x2, y2 (inherited via GetRect)
    GetRect() image.Rectangle
    SetRect(x1, y1, x2, y2 int)
    Draw(*Buffer)
    sync.Locker
}

// A widget that also consumes input implements EventHandler.
type EventHandler interface {
    HandleEvent(Event) bool   // return true => event consumed
}

type Buffer struct {
    image.Rectangle
    Cells []Cell
}

type Cell struct {
    Rune  rune
    Style Style
}

type Style struct {
    Fg       Color
    Bg       Color
    Modifier AttrMask
}
```

## The Three Building Blocks

### 1. `Block` — every widget embeds it

`Block` draws the border, title, padding, and background. Almost every widget
in `widgets/` starts with:

```go
type Paragraph struct {
    Block                 // <-- embedded
    Text         string
    TextStyle    Style
    // ...
}

func (p *Paragraph) Draw(buf *ui.Buffer) {
    p.Block.Draw(buf)     // delegate border/title/padding first
    // ... then draw content into buf
}
```

**Lesson**: build widgets by embedding `Block`, not by re-implementing borders.
You inherit `Title`, `TitleRight`, `TitleBottom`, `Border`, `BorderRounded`,
`Padding`, `Inner` rectangle math, and consistent styling.

### 2. `Buffer` + `Cell` — the rendering surface

`Buffer` is a flat slice of cells sized to the union of all drawables.
Use `buf.SetString(s, style, image.Pt(x, y))` to write text width-aware
(thanks to `go-runewidth`), or `buf.SetCell(Cell, image.Pt(x, y))` for a
single rune.

You normally do **not** construct buffers yourself — `Render()` does that.
But advanced widgets (e.g. `Canvas`, `Image`) need direct buffer access.

### 3. `Application` / `Backend` — the host

- `Backend` wraps `tcell.Screen` plus an `Init/Close/Render/PollEvents`
  surface. Two constructors:
  - `DefaultBackend` — global singleton used by `ui.Init/Render/Close`.
  - `ui.NewBackend(&ui.InitConfig{CustomTTY: ...})` — isolated backend per
    instance, **required for SSH servers** and tests.
- `Application` wraps a `Backend`, holds a root widget + focus widget, and
  runs the main event loop. This is the recommended entry point for new
  apps.

## Two-Phase Render

```go
// Render pipeline (paraphrased)
func Render(items ...Drawable) {
    buf := NewBuffer(calculateBounds(items))   // phase 1: size
    for _, d := range items {
        d.Lock()
        d.Draw(buf)                            // phase 2: fill
        d.Unlock()
    }
    DefaultBackend.renderBuffer(buf)           // phase 3: flush to tcell
}
```

**Why this matters**:
- One terminal write per frame → no flicker, no partial renders.
- Widgets can overlap (the last draw wins).
- Trivially extensible: add a new widget by implementing `Drawable`.

## Event Flow (Application)

```
tcell.NewEventKey/Mouse/Resize
        │
        ▼
canonical conversion  ──▶  Event{Type, ID, Payload}
        │
        ▼
Application.handleEvent(e)
        │
        ├── ResizeEvent → handleResize (SetRect + Clear + Render)
        │
        ├── KeyboardEvent → focus.HandleEvent(e) → if !handled root.HandleEvent(e)
        ├── MouseEvent    → root.HandleEvent(e)
        │
        ├── if !handled && (e.ID == "q" || e.ID == "<C-c>") → stop
        │
        └── otherwise: Clear + Render(root)
```

Key points:
- Keyboard is dispatched to the **focused widget first**, then bubbles to root.
- Mouse always targets the root.
- `q` and `<C-c>` are reserved by default. `<Escape>` is **not** by default —
  handle it yourself if needed.

## Concurrency

Every `Drawable` embeds `sync.Mutex`. Pattern:

```go
d.Lock()
d.SomeField = x
d.Unlock()
```

The `Application.run` loop holds the root's lock briefly during `SetRect`
to avoid races with concurrent `Draw` calls. If your widget mutates from
a goroutine, lock it the same way.

## What to Use Next

| Need                         | Use                                      |
| ---------------------------- | ---------------------------------------- |
| Shared type overview         | This document's "Core Types" section    |
| Custom widget template       | This document's `Block` embedding pattern |
| Event loop behavior          | This document's "Event Flow" section    |
| Key names                    | `references/events.md`                   |
| SSH backend pattern          | `references/ssh-server.md`               |