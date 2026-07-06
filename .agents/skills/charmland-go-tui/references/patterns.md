# Patterns and Pitfalls

## Project Shape

For a single TUI binary, keep the entry point simple:

```text
go.mod
main.go
internal/app/model.go
internal/app/update.go
internal/app/view.go
```

For multiple commands, use `cmd/<name>/main.go` and shared packages under `internal/`.

## Model Design

Good model fields:

- Domain state needed to render the screen.
- Current route/screen/mode.
- Window dimensions.
- Child Bubbles component models.
- Current loading/error state.

Avoid:

- Open network connections in the model when a command can own the operation.
- Global mutable UI state.
- Deriving view-only strings in `Update` unless rendering is expensive, such as Markdown.

## Screen Routing

For small apps, a simple enum is enough:

```go
type screen int

const (
	screenHome screen = iota
	screenDetails
	screenForm
)
```

In `Update`, dispatch by screen after global keys. In `View`, switch on screen and compose screen-specific content.

## Async Work

Do not block in `Update`:

```go
func fetchItems() tea.Msg {
	items, err := api.FetchItems(context.Background())
	if err != nil {
		return errMsg{err}
	}
	return itemsLoadedMsg{items}
}

func (m model) Init() tea.Cmd {
	return fetchItems
}
```

For cancelable work, carry a `context.Context` into the command or use `tea.WithContext` at program creation.

## Layout Calculation

Store dimensions from `tea.WindowSizeMsg`. Derive layout sizes from them:

```go
header := headerStyle.Render("Title")
footer := footerStyle.Render("q quit")
bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
if bodyHeight < 0 {
	bodyHeight = 0
}
m.viewport.SetHeight(bodyHeight)
```

## Forms vs Components

Use Huh for linear or grouped data collection. Use Bubbles directly when the form is part of a larger app screen or needs custom interactions across panes.

## Markdown

Use Glamour for Markdown and put rendered output into a Bubbles `viewport`. Re-render when width or Markdown source changes.

## Common Pitfalls

- Returning `string` from an app model's `View` in Bubble Tea v2. Return `tea.View` and wrap strings with `tea.NewView`.
- Forgetting to store the child model returned by a Bubbles component `Update`.
- Sending all keys to all components instead of respecting focus.
- Blocking in `Update` with network or filesystem work.
- Recreating expensive renderers every frame for large Markdown content.
- Printing logs to stdout while the TUI is running.
- Ignoring terminal resize messages.
- Applying fixed widths that break on narrow terminals.

## Verification Checklist

- `go test ./...` or at least `go test` for touched packages.
- `go vet ./...` if the project already uses it.
- `go run .` for small apps when feasible.
- Manual terminal checks for resize, quit keys, narrow width, and failure states.
