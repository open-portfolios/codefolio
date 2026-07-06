# Layout: SetRect / Grid / Flex

gotui provides **three layout strategies**. Pick the simplest one that
fits — don't reach for Grid when SetRect is enough.

## 1. Absolute Positioning — `SetRect`

Every widget has:

```go
w.SetRect(x1, y1, x2, y2 int)   // exclusive end (image.Rectangle)
```

Coordinates are in terminal cells, with `(0, 0)` at top-left and `x2/y2`
**exclusive** (so a 10-wide widget starting at `x=0` ends at `x=10`).

**Use when**:
- You have a single widget filling the screen.
- You're building a custom layout (overlapping widgets, modal layers).
- You're prototyping.

```go
header := widgets.NewParagraph()
header.SetRect(0, 0, termW, 3)

body := widgets.NewParagraph()
body.SetRect(0, 3, termW, termH)

footer := widgets.NewParagraph()
footer.SetRect(0, termH-1, termW, termH)
```

**On resize** (Application OO API does this automatically; otherwise
listen for `<Resize>`):

```go
case "<Resize>":
    r := e.Payload.(ui.Resize)
    header.SetRect(0, 0, r.Width, 3)
    body.SetRect(0, 3, r.Width, r.Height-1)
    footer.SetRect(0, r.Height-1, r.Width, r.Height)
    ui.Clear()
    ui.Render(header, body, footer)
```

## 2. Ratio Grid — `ui.NewGrid()`

The most ergonomic layout for dashboards. Ratios are normalized automatically
— you can pass any positive float.

```go
grid := ui.NewGrid()
grid.Set(
    ui.NewRow(1.0/8,                  // row height = 12.5%
        ui.NewCol(1.0, header),       // 1 col, full width
    ),
    ui.NewRow(3.0/8,                  // row height = 37.5%
        ui.NewCol(1.0/2, leftPanel),  // left half
        ui.NewCol(1.0/2,             // right half
            ui.NewRow(1.0/2, topRight),
            ui.NewRow(1.0/2, botRight),
        ),
    ),
    ui.NewRow(4.0/8,                  // row height = 50%
        ui.NewCol(1.0, footer),
    ),
)
grid.SetRect(0, 0, termW, termH)
ui.Render(grid)
```

**Rules**:
- Rows are sized by ratio of their `h` to the sum of all row ratios.
- Cols are sized by ratio of their `w` to the sum of all col ratios in the row.
- Nested `ui.NewRow` inside a `ui.NewCol` works — any `Drawable` can hold a
  sub-grid.

**Always check the ratios sum to 1.0 per axis** for clarity, though the
library normalizes automatically.

**Use when**: dashboards, multi-panel monitoring, anything > 3 widgets.

## 3. Flex — `widgets.NewFlex()`

A row-or-column container with **fixed-size** and **proportional** items.

```go
flex := widgets.NewFlex()
flex.SetDirection(ui.FlexColumn)         // or FlexRow

// Fixed: takes exactly 20 cells regardless of container size.
// Proportion: takes remaining space proportional to its weight.
flex.AddItem(sidebar, 20, 0, false)      // fixed 20
flex.AddItem(main,    0, 1, true)        // flexible, focusable
flex.AddItem(extra,   0, 2, false)       // takes 2/3 of remaining

flex.SetRect(0, 0, termW, termH)
ui.Render(flex)
```

**Args to `AddItem`**:
1. `item` — the Drawable.
2. `fixedSize` — fixed cell count (0 = proportional).
3. `proportion` — relative weight when `fixedSize == 0`.
4. `focusable` — receive focus via `Tab`/click.

**Use when**: sidebars, toolbars, status bars, anything with one prominent
flexible region.

## Decision Tree

```
1 widget filling screen?
  └─ SetRect

2-3 panels side-by-side?
  └─ Flex row/column

4+ panels, complex nested layout, ratios?
  └─ Grid

Modal over existing content?
  └─ SetRect (absolute, on top of Grid)
```

## Nested Grid Pattern

The cleanest dashboard pattern is **Grid of Grids**:

```go
header := widgets.NewParagraph()
sidebar := widgets.NewList()
content := widgets.NewParagraph()
footer := widgets.NewParagraph()

grid := ui.NewGrid()
grid.Set(
    ui.NewRow(1.0/12,
        ui.NewCol(1.0, header),
    ),
    ui.NewRow(10.0/12,
        ui.NewCol(1.0/4, sidebar),
        ui.NewCol(3.0/4, content),
    ),
    ui.NewRow(1.0/12,
        ui.NewCol(1.0, footer),
    ),
)
grid.SetRect(0, 0, termW, termH)
```

This resizes cleanly: just call `grid.SetRect` again on `<Resize>`.

## Padding & Inner Rectangle

Every widget that embeds `Block` automatically computes `Block.Inner`:

```go
type Block struct {
    image.Rectangle            // outer rect (set by SetRect)
    Inner       image.Rectangle // content area, after borders+padding+title
    Border      bool
    BorderRounded bool
    Padding     image.Rectangle // left/top/right/bottom inset
    Title       string
    TitleRight  string
    TitleBottom string
}
```

When implementing a custom widget, draw into `b.Inner`, not the outer
rectangle — this guarantees you don't overlap your own border or title.

## Modals / Overlays

There is **no** built-in z-order manager. To layer a modal over a Grid:

```go
modal := widgets.NewModal()
modal.SetRect(20, 5, 60, 20)

ui.Render(grid, modal)   // order matters: modal drawn last → on top
```

Both `grid` and `modal` get `SetRect` independently; `Render` draws them
both into the same buffer. The cells where they overlap are overwritten by
whichever draws last.

## Anti-Patterns

- ❌ Calling `SetRect` on the root of a Grid — Grid manages child rects.
- ❌ Mixing Flex + Grid children of the same parent — pick one model.
- ❌ Hardcoding widths/heights that don't sum — use ratios.