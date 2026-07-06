# Widget Catalog

All widgets below ship in the `widgets` subpackage.

## Quick Picker

| I want to…                                 | Use                            |
| ------------------------------------------ | ------------------------------ |
| Show static text                           | `Paragraph`                    |
| Show scrollable text                       | `Paragraph` + wrap             |
| Show a clickable, scrollable list of items | `List`                         |
| Show a multi-column table                  | `Table`                        |
| Show a horizontal / vertical tab bar       | `Tabs` (`TabPane`)             |
| Show an editable text field                | `Input`                        |
| Show a multi-line text editor              | `TextArea`                     |
| Show a clickable button                    | `Button`                       |
| Show a toggle / checkbox                   | `Checkbox`                     |
| Show a progress bar                        | `Gauge`                        |
| Show a single-line progress                | `LineGauge`                    |
| Show a sparkline (mini line chart)         | `Sparkline` / `SparklineGroup` |
| Show a line chart                          | `Plot`                         |
| Show a bar chart                           | `BarChart` / `StackedBarChart` |
| Show a step chart                          | `StepChart`                    |
| Show a pie chart                           | `PieChart`                     |
| Show a donut chart                         | `DonutChart`                   |
| Show a funnel chart                        | `FunnelChart`                  |
| Show a radar chart                         | `RadarChart`                   |
| Show a heatmap                             | `Heatmap`                      |
| Show a treemap                             | `Treemap`                      |
| Show a tree                                | `Tree`                         |
| Show a calendar                            | `Calendar`                     |
| Show an animated spinner                   | `Spinner`                      |
| Show a logo / image                        | `Image` / `Canvas`             |
| Show a modal dialog                        | `Modal`                        |
| Show a scrollbar next to another widget    | `Scrollbar`                    |

## Display Widgets

### `Paragraph`

The workhorse. Multi-line text, supports inline style DSL.

```go
p := widgets.NewParagraph()
p.Title = "About"
p.Text = "Hello [world](fg:red,mod:bold)!"
p.BorderRounded = true
```

### `List`

Scrollable single-column list with optional selection.

```go
l := widgets.NewList()
l.Title = "Files"
l.Rows = []string{"a.txt", "b.txt", "c.txt"}
l.SelectedRow = 0
// Listen for <Up>/<Down>/<MouseWheelUp>/<MouseWheelDown> in HandleEvent.
```

### `Table`

Multi-column table with optional row separators.

```go
t := widgets.NewTable()
t.Rows = [][]string{
    {"Name", "Size", "Modified"},
    {"a.txt", "12K", "2026-01-01"},
}
t.ColumnWidths = []int{20, 10, 20}
t.FillRow = true
```

### `Tabs`

```go
tabs := widgets.NewTabPane("Overview", "Details", "Logs")
tabs.SetRect(0, 0, 60, 3)
tabs.ActiveTabIndex = 0
// Tabs.BorderLeft/Right lets you connect to the active panel's border.
```

### `Scrollbar`

Attach to any widget with scrollable content. Use together with `List` or
`Paragraph` for a scrollbar + content layout.

## Input Widgets

### `Input` (single-line)

```go
in := widgets.NewInput()
in.Placeholder = "Search…"
in.OnSubmit = func(text string) { /* handle Enter */ }
```

### `TextArea` (multi-line)

Like `Input` but supports newlines, optional line numbers, undo.

### `Button`

```go
b := widgets.NewButton("Submit")
b.OnClick = func() { /* ... */ }
```

### `Checkbox`

```go
c := widgets.NewCheckbox()
c.Label = "I agree"
c.Checked = false
```

## Progress & Charts

### `Gauge` (block-style)

```go
g := widgets.NewGauge()
g.Percent = 75
g.BarColor = ui.ColorGreen
g.Label = "75%"
```

### `LineGauge` (single-line)

```go
lg := widgets.NewLineGauge()
lg.Percent = 60
lg.BarRune = '▰'
lg.BarRuneEmpty = '▱'
lg.LineColor = ui.ColorYellow
```

### `Sparkline` / `SparklineGroup`

```go
sl := widgets.NewSparkline()
sl.Data = []float64{1, 2, 3, 5, 8, 13}
sl.MaxVal = 20
sl.LineColor = ui.ColorGreen

slg := widgets.NewSparklineGroup(sl1, sl2, sl3) // stacked
```

### `Plot` (multi-series line)

```go
plot := widgets.NewPlot()
plot.Data = [][]float64{series1, series2}
plot.LineColors[0] = ui.ColorGreen
plot.Marker = widgets.MarkerDot
```

### `BarChart`

```go
bc := widgets.NewBarChart()
bc.Data = []float64{3, 2, 5, 3, 9, 5}
bc.Labels = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
bc.BarColors = []ui.Color{ui.ColorBlue, ui.ColorLightCyan}
```

### `PieChart`

```go
pc := widgets.NewPieChart()
pc.Data = []float64{10, 20, 30, 40}
pc.Colors = []ui.Color{ui.ColorRed, ui.ColorYellow, ui.ColorGreen, ui.ColorBlue}
pc.LabelFormatter = func(i int, v float64) string {
    return fmt.Sprintf("%.0f%%", v)
}
```

### Other charts

`StackedBarChart`, `StepChart`, `RadarChart`, `FunnelChart`, `Heatmap`,
`Treemap` — all share the chart pattern: `Data []float64`, `Colors []Color`,
optional `Labels []string`.

## Spinner

```go
sp := widgets.NewSpinner()
sp.Frames = widgets.SpinnerDots
sp.Label = "Loading…"
// Drive: sp.Advance() + Render in a ticker.
```

See `assets/spinner-frames.md` for 17 ready-made frame sets.

## Special

### `Canvas` (braille drawing)

Plot pixel-style dots that get rendered into braille characters
(2× density). Useful for low-resolution images inside the terminal.

```go
c := widgets.NewCanvas()
c.SetPoint(x, y)        // set one dot
c.SetLine(x1, y1, x2, y2) // Bresenham line
c.SetRect(0, 0, 40, 20)
c.Draw(buf)              // requires direct buffer access
```

### `Image`

Render an `image.Image` (PNG, JPEG, GIF frame) into a Block using half-block
characters for 2× vertical density.

### `Modal`

Layered dialog that draws on top of everything else.

```go
m := widgets.NewModal()
m.SetRect(20, 5, 60, 20)
m.Title = "Confirm"
m.Text = "Are you sure?"

// Order in Render() controls z-order:
ui.Render(grid, modal)  // modal drawn last → on top
```

### `Tree` (collapsible)

```go
tr := widgets.NewTree()
tr.SetNodes([]*widgets.TreeNode{
    {Text: "root", Children: []*widgets.TreeNode{
        {Text: "child A"},
        {Text: "child B"},
    }},
})
```

### `Calendar`

```go
cal := widgets.NewCalendar()
cal.HeaderStyle = ui.NewStyle(ui.ColorCyan)
cal.CursorStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow)
```

### `Logo`

> **Note**: This widget renders the `gotui` brand ASCII logo. It is
> brand-specific. Avoid using it in your own application — replace with
> a custom `Paragraph` containing your own ASCII art, or remove the title.

## Common Field Reference

Most widgets share these:

```go
Title             string
TitleRight        string
TitleBottom       string
TitleStyle        ui.Style
TitleAlignment    ui.Alignment  // AlignLeft | AlignCenter | AlignRight
Border            bool           // default true
BorderRounded     bool
BorderStyle       ui.Style
Padding           image.Rectangle
Inner             image.Rectangle // computed, read-only
```

## Combining Widgets

Typical patterns:

- **Dashboard**: `Grid` of `SparklineGroup`, `Gauge`, `BarChart`, `List`, `Gauge`.
- **Form**: `Grid` of `Input`, `Input`, `Button`.
- **Detail view**: `Tabs` over multiple `Paragraph` panels.
- **Monitor**: single `SparklineGroup` filling the screen.
- **Menu**: `List` with `OnSelectionChanged` handler.

When in doubt, start with `Paragraph` + `List` + `Grid`. You can always
swap in richer widgets later.