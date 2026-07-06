# Lip Gloss v2 Reference

Lip Gloss is the styling and layout layer for Charm TUIs. It provides CSS-like immutable `Style` values, layout helpers, and structural renderers.

## Style Basics

```go
import "charm.land/lipgloss/v2"

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	Padding(0, 1)

title := titleStyle.Render("Dashboard")
```

`Style` is a value type. Methods return modified copies, so package-level styles are safe to derive from:

```go
base := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
selected := base.Bold(true).Foreground(lipgloss.Color("212"))
```

## Common Style Methods

- Text: `Bold`, `Italic`, `Underline`, `Strikethrough`, `Faint`, `Reverse`.
- Color: `Foreground`, `Background`, `UnderlineColor`.
- Box model: `Padding`, `Margin`, `Border`, `BorderStyle`, `BorderForeground`.
- Dimensions: `Width`, `Height`, `MaxWidth`, `MaxHeight`.
- Alignment: `Align`, `AlignHorizontal`, `AlignVertical`.
- Transform: `Transform(func(string) string)`.

## Borders

Useful border factories:

- `lipgloss.NormalBorder()`
- `lipgloss.RoundedBorder()`
- `lipgloss.ThickBorder()`
- `lipgloss.DoubleBorder()`
- `lipgloss.HiddenBorder()`
- `lipgloss.ASCIIBorder()`
- `lipgloss.MarkdownBorder()`

Example:

```go
panel := lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(1, 2).
	Render(content)
```

## Layout Helpers

Use `JoinHorizontal` and `JoinVertical` for simple composition:

```go
body := lipgloss.JoinHorizontal(
	lipgloss.Top,
	sidebarStyle.Render(sidebar),
	mainStyle.Render(main),
)
```

Use `Place`, `PlaceHorizontal`, and `PlaceVertical` to align content inside a fixed area:

```go
centered := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
```

Measure rendered content with:

```go
w := lipgloss.Width(content)
h := lipgloss.Height(content)
w, h = lipgloss.Size(content)
```

## Structural Subpackages

Use these when static rendering is enough and Bubble Tea interactivity is not needed:

- `charm.land/lipgloss/v2/table` for styled static tables.
- `charm.land/lipgloss/v2/list` for nested static lists.
- `charm.land/lipgloss/v2/tree` for tree output.

For interactive tables/lists, use Bubbles `table` or `list` instead.

## Integration with Bubble Tea

- Build strings with Lip Gloss inside `View`.
- Return `tea.NewView(renderedString)` from Bubble Tea v2 app models.
- Recompute width-dependent layout from dimensions stored after `tea.WindowSizeMsg`.
- Keep style values outside `Update`; only dimensions and app state belong in the model.
