# Events & Key Bindings

## Event Types

```go
const (
    KeyboardEvent EventType = iota
    MouseEvent
    ResizeEvent
)

type Event struct {
    Type    EventType
    ID      string    // canonical name like "<C-c>", "<Up>", "<MouseLeft>"
    Payload interface{}
}
```

| Type            | Payload                    | ID examples                                             |
| --------------- | -------------------------- | ------------------------------------------------------- |
| `KeyboardEvent` | `*tcell.EventKey`          | `"a"`, `"<C-c>"`, `"<Up>"`, `"<F1>"`, `"<M-x>"`         |
| `MouseEvent`    | `ui.Mouse{X, Y, Drag}`     | `"<MouseLeft>"`, `"<MouseWheelUp>"`, `"<MouseRelease>"` |
| `ResizeEvent`   | `ui.Resize{Width, Height}` | `"<Resize>"`                                            |

## Canonical Key ID Table

The library translates raw `tcell.Key` constants into stable string IDs so
your widget code never depends on `tcell` directly.

### Letters & numbers

| Input            | ID                                   |
| ---------------- | ------------------------------------ |
| `a`–`z`, `A`–`Z` | the rune itself (e.g. `"a"`, `"Z"`)  |
| `0`–`9`          | the rune itself                      |
| Space            | `" "` or `"<Space>"` (both accepted) |
| Tab              | `"<Tab>"`                            |
| Enter            | `"<Enter>"`                          |
| Escape           | `"<Escape>"`                         |
| Backspace        | `"<Backspace>"`                      |

### Arrow keys

| Key         | ID          |
| ----------- | ----------- |
| Up arrow    | `"<Up>"`    |
| Down arrow  | `"<Down>"`  |
| Left arrow  | `"<Left>"`  |
| Right arrow | `"<Right>"` |

### Navigation

| Key      | ID             |
| -------- | -------------- |
| Home     | `"<Home>"`     |
| End      | `"<End>"`      |
| PageUp   | `"<PageUp>"`   |
| PageDown | `"<PageDown>"` |
| Insert   | `"<Insert>"`   |
| Delete   | `"<Delete>"`   |

### Ctrl + letter

| Key    | ID                              |
| ------ | ------------------------------- |
| Ctrl-A | `"<C-a>"`                       |
| Ctrl-B | `"<C-b>"`                       |
| Ctrl-C | `"<C-c>"` *(reserved for quit)* |
| Ctrl-D | `"<C-d>"`                       |
| ...    | ...                             |
| Ctrl-Z | `"<C-z>"`                       |

### Alt + letter

| Key   | ID        |
| ----- | --------- |
| Alt-X | `"<M-x>"` |

### Function keys

`"<F1>"` through `"<F12>"`.

### Mouse

| Action            | ID                   |
| ----------------- | -------------------- |
| Left click / drag | `"<MouseLeft>"`      |
| Middle click      | `"<MouseMiddle>"`    |
| Right click       | `"<MouseRight>"`     |
| Wheel up          | `"<MouseWheelUp>"`   |
| Wheel down        | `"<MouseWheelDown>"` |
| Button release    | `"<MouseRelease>"`   |

Mouse events carry `Payload` of type `ui.Mouse`:

```go
type Mouse struct {
    X, Y int
    Drag bool
}
```

### Resize

`Payload` is `ui.Resize`:

```go
type Resize struct {
    Width  int
    Height int
}
```

## Reserved / Default Handled

When using `Application.Run()` (the OO API), these are reserved:

| ID      | Action |
| ------- | ------ |
| `q`     | Quit   |
| `<C-c>` | Quit   |

`<Escape>` is **not** reserved — handle it yourself if you want it to
back-out of a modal.

## Dispatch Order (Application)

```
KeyboardEvent
   1. focus.HandleEvent(e)
   2. if !handled → root.HandleEvent(e)
   3. if !handled → global quit check (q / <C-c>)
   4. Render(root)

MouseEvent
   1. root.HandleEvent(e)
   2. Render(root)

ResizeEvent
   1. handleResize: SetRect(root, w, h), Clear, Render(root)
```

## Writing `HandleEvent`

```go
func (w *MyWidget) HandleEvent(e ui.Event) bool {
    switch e.Type {
    case ui.KeyboardEvent:
        switch e.ID {
        case "j", "<Down>":
            w.cursor++
            return true
        case "k", "<Up>":
            w.cursor--
            return true
        case "<C-d>":
            w.cursor += 10
            return true
        }
    case ui.MouseEvent:
        m := e.Payload.(ui.Mouse)
        if e.ID == "<MouseLeft>" && m.Y == w.Inner.Min.Y + w.cursor {
            w.activate(m.Y)
            return true
        }
    }
    return false   // let other handlers try
}
```

**Conventions**:
- Return `true` to mark the event consumed.
- Mouse coordinates are **terminal cell** coordinates (not pixels).
- Always check `e.Type` first — different types have different payloads.

## Polling (Functional API)

```go
for e := range ui.PollEvents() {
    switch e.ID {
    case "q", "<C-c>":
        return
    case "<Resize>":
        r := e.Payload.(ui.Resize)
        grid.SetRect(0, 0, r.Width, r.Height)
        ui.Clear()
        ui.Render(grid)
    case "<MouseWheelUp>":
        list.ScrollUp()
        ui.Render(grid)
    }
}
```

## Checking `tcell.ModAlt` etc.

If you need the raw tcell event (e.g. to inspect shift state), use
`e.Payload.(*tcell.EventKey).Modifiers()`:

```go
if ke, ok := e.Payload.(*tcell.EventKey); ok {
    if ke.Modifiers()&tcell.ModShift != 0 {
        // shifted key
    }
}
```

Avoid this when you can — the canonical IDs already encode Alt as
`"<M-x>"`.