# Styling & Theming

gotui styling is intentionally lightweight: **flat `Style{Fg, Bg, Modifier}`**
plus an optional inline DSL parser and a global theme singleton. No CSS
selectors, no styled-chain like lipgloss.

## The `Style` Type

```go
type Style struct {
    Fg       Color
    Bg       Color
    Modifier AttrMask
}
```

Construct via:

```go
s := ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
// or just Fg:
s := ui.NewStyle(ui.ColorCyan)
// or empty (inherits terminal default):
s := ui.StyleClear
```

## Built-in Colors

### ANSI 16

```go
ui.ColorBlack
ui.ColorRed
ui.ColorGreen
ui.ColorYellow
ui.ColorBlue
ui.ColorMagenta
ui.ColorCyan
ui.ColorWhite
ui.ColorLightGray       (a.k.a. ColorDefault in some terminals)
ui.ColorDarkGray

ui.ColorLightRed
ui.ColorLightGreen
ui.ColorLightYellow
ui.ColorLightBlue
ui.ColorLightMagenta
ui.ColorLightCyan
ui.ColorLightGray       // light gray ANSI color
```

### Special

```go
ui.ColorDefault   // inherit terminal default
ui.ColorClear     // explicitly no color (preserves underlying cell)
```

### TrueColor (24-bit)

```go
ui.NewRGBColor(r, g, b uint8) // returns ui.Color
// e.g.
ui.NewRGBColor(255, 100, 50)
```

Hex parser:

```go
ui.HexToColor("#ff6432")
```

## Modifiers (SGR attributes)

```go
ui.ModifierBold
ui.ModifierUnderline     // Note: removed in tcell v3 in some contexts;
                         // Block.drawBorder uses Style.Underline() directly
ui.ModifierReverse
ui.ModifierBlink
ui.ModifierDim
```

Combine with bitwise OR:

```go
ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold|ui.ModifierUnderline)
```

## The Global Theme

`ui.Theme` is a package-level mutable singleton:

```go
var Theme = RootTheme{
    Default: ui.Style{Fg: ui.ColorWhite, Bg: ui.ColorDefault},
    Block:   ui.BlockTheme{Default: ui.Style{...}, Title: ..., Border: ...},
    Paragraph: ui.ParagraphTheme{Text: ...},
    BarChart:  ui.BarChartTheme{...},
    // ... one theme block per widget
}
```

Every widget's `New*()` constructor seeds its style fields from `ui.Theme`.
To re-skin globally:

```go
ui.Theme.Block.Title.Fg = ui.ColorCyan
ui.Theme.Paragraph.Text.Fg = ui.ColorWhite
```

**Caveat**: global state. Acceptable for apps; problematic if you embed the
library in something that wants multiple themes.

## Per-Widget Styling (the common path)

Each widget exposes a handful of style fields. Set them before rendering:

```go
p := widgets.NewParagraph()
p.TextStyle.Fg = ui.ColorWhite
p.TitleStyle = ui.NewStyle(ui.ColorLightCyan, ui.ColorClear, ui.ModifierBold)
p.BorderStyle.Fg = ui.ColorLightCyan
p.BorderRounded = true
```

### Common style fields

| Widget                | Fields                                                       |
| --------------------- | ------------------------------------------------------------ |
| `Block` (all widgets) | `BorderStyle`, `TitleStyle`, `Bg`, `Border`, `BorderRounded` |
| `Paragraph`           | `TextStyle`                                                  |
| `List`                | `TextStyle`, `SelectedStyle`                                 |
| `Table`               | `TextStyle`, `SelectedStyle`, `RowStyles`, `ColumnStyles`    |
| `Sparkline`           | `LineColor`, `TitleStyle`, `BorderStyle`                     |
| `BarChart`            | `BarColors`, `NumStyles`, `LabelStyles`                      |
| `PieChart`            | `Colors`                                                     |
| `Gauge`               | `BarColor`, `LabelStyle`                                     |

## Inline Style DSL

Paragraphs, Lists, and Tables can parse inline style markers at draw time:

```
[normal text](fg:red,bg:blue,mod:bold) [more text](fg:green)
```

Grammars:

```
[text]                     -> text in defaultStyle
[text](fg:COLOR)           -> text in COLOR (foreground only)
[text](bg:COLOR)           -> text with background
[text](mod:BOLD)           -> text with modifier (BOLD|UNDERLINE|REVERSE)
[text](fg:COLOR,bg:COLOR,mod:MOD1|MOD2)
```

Multiple style segments can be concatenated in one string. The parser is in
`ParseStyles(s string, defaultStyle Style) []Cell`.

```go
p := widgets.NewParagraph()
p.Text = "Hello [world](fg:red,mod:bold)! Click [here](fg:cyan,mod:underline) for more."
```

**Cost**: parsed every `Draw()`. Cache the result if the text never changes
(advanced optimization — usually unnecessary).

## Gradients

For border / text color gradients:

```go
p.BorderGradient = ui.Gradient{
    Enabled:   true,
    Start:     ui.ColorRed,
    End:       ui.ColorBlue,
    Direction: ui.GradientHorizontal,
}

// Multi-stop:
g := ui.GenerateMultiGradient(
    []ui.Color{ui.ColorRed, ui.ColorYellow, ui.ColorGreen},
    steps,
)
```

For programmatic use:

```go
// Linear gradient between two colors, n steps.
colors := ui.GenerateGradient(ui.ColorRed, ui.ColorBlue, 16)

// Interpolate between two colors at t in [0,1].
c := ui.InterpolateColor(ui.ColorRed, ui.ColorBlue, 0.5)
```

Use cases:
- Animated progress bars (regenerate each tick).
- Heatmap-style data display.
- Branded headers.

## Border Sets

```go
p.BorderRounded = true     // round corners (UTF-8 box-drawing)
p.BorderType   = ui.BorderThick   // preset (Plain|Round|Thick|Double|Hidden)
```

Border sets are structs of box-drawing runes:

```go
type BorderSet struct {
    Top, Bottom         rune
    Left, Right         rune
    TopLeft, TopRight   rune
    BottomLeft, BottomRight rune
}
```

Use `ui.BorderSet*` presets, or define your own.

## Anti-Patterns

- ❌ Mutating `ui.Theme` inside a hot loop — it's a global. Cache.
- ❌ Calling `ParseStyles` on every render with static text — precompute.
- ❌ Mixing RGB and ANSI colors in a single gradient — RGB → ANSI quantization
  is lossy and uneven.