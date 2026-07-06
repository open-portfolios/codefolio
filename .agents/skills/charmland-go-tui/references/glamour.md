# Glamour Reference

Glamour renders Markdown to ANSI-styled terminal output. Use it for help screens, README previews, release notes, docs, rich descriptions, and Markdown content inside Bubble Tea apps.

## Basic Rendering

```go
import "charm.land/glamour/v2"

out, err := glamour.Render("# Hello\n\nThis is **Markdown**.", "dark")
if err != nil {
	return err
}
fmt.Print(out)
```

## Renderer with Options

```go
r, err := glamour.NewTermRenderer(
	glamour.WithStandardStyle("dark"),
	glamour.WithWordWrap(80),
	glamour.WithEmoji(),
)
if err != nil {
	return err
}

out, err := r.Render(markdown)
```

Useful options:

- `WithStandardStyle(name)` for built-in styles.
- `WithWordWrap(width)` for terminal width.
- `WithStylesFromJSONFile(path)` or `WithStylesFromJSONBytes(data)` for custom themes.
- `WithPreservedNewLines()` when source line breaks matter.
- `WithEmoji()` for emoji shortcodes.
- `WithBaseURL(url)` for resolving relative links.

## Built-in Styles

- `dark`
- `light`
- `dracula`
- `tokyo-night`
- `pink`
- `ascii`
- `notty`

## Bubble Tea Integration

Render Markdown when the source or width changes, not every frame if the input is large.

```go
case tea.WindowSizeMsg:
	m.width = msg.Width
	m.renderMarkdown()

case markdownLoadedMsg:
	m.markdown = msg.body
	m.renderMarkdown()
```

Store the rendered string in the model and display it in a `viewport` when content can exceed the screen height.

```go
func (m *model) renderMarkdown() {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(max(20, m.width-4)),
	)
	if err != nil {
		m.err = err
		return
	}
	m.rendered, m.err = r.Render(m.markdown)
	m.viewport.SetContent(m.rendered)
}
```

## Color Output

Glamour v2 is intentionally pure. Terminal color adaptation is generally handled by the caller, often through Lip Gloss output functions in non-Bubble Tea contexts. Inside Bubble Tea, return rendered content through the Bubble Tea renderer.
