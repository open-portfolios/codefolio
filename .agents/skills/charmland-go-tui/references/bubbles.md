# Bubbles v2 Reference

Bubbles is the official component library for Bubble Tea. Components are model-like values with `Init`, `Update`, and `View` methods, but their `View()` methods usually return `string`. In a Bubble Tea v2 app, compose those strings and return `tea.NewView(...)` from the parent app model.

## Component Catalog

| Package      | Use                                                                   |
| ------------ | --------------------------------------------------------------------- |
| `textinput`  | Single-line input, passwords, suggestions, validation.                |
| `textarea`   | Multi-line text editor.                                               |
| `viewport`   | Scrollable content area.                                              |
| `list`       | Full-featured item browser with filtering, pagination, help, spinner. |
| `table`      | Navigable table with selected row and viewport scrolling.             |
| `spinner`    | Animated activity indicator.                                          |
| `progress`   | Static or animated progress bar.                                      |
| `paginator`  | Pagination logic and UI.                                              |
| `filepicker` | Filesystem navigation and selection.                                  |
| `help`       | Auto-render help from key bindings.                                   |
| `key`        | Key binding definitions and matching.                                 |
| `timer`      | Countdown timer.                                                      |
| `stopwatch`  | Count-up timer.                                                       |
| `cursor`     | Virtual cursor, mostly used by inputs.                                |

## Composition Pattern

```go
import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
)

type model struct {
	input textinput.Model
}

func newModel() model {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	return model{input: ti}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	return tea.NewView(m.input.View() + "\n")
}
```

If the exact constructor or field names differ in the target project, follow the installed version and existing project code. Verify against the local package source when uncertain.

## Key Bindings

Use the `key` package for app-level bindings:

```go
import "charm.land/bubbles/v2/key"

type keyMap struct {
	Quit key.Binding
	Save key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Save: key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
}

case tea.KeyPressMsg:
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Save):
		return m, saveCmd(m)
	}
```

## Choosing Components

- Use `viewport` for long text, logs, previews, Markdown output, and scrollable panes.
- Use `textinput` for short filters/search/forms; use `textarea` for multi-line editing.
- Use `list` when items need filtering, selection, pagination, item delegates, or built-in help.
- Use `table` when data is naturally rows and columns and needs keyboard navigation.
- Use `spinner` for unknown duration work; use `progress` when percentage is known.
- Use `help` when you maintain a keymap and want discoverable controls.

## Focus Rules

Interactive components usually need focus. Maintain focus explicitly when composing several components:

- Track the active component in the parent model.
- Forward key messages only to the active component unless global shortcuts should win.
- Call child `Focus()` and `Blur()` when switching.

## Layout Rules

- Update component width/height after `tea.WindowSizeMsg`.
- Keep header/footer heights fixed or measured with `lipgloss.Height`.
- Give remaining height to `viewport`, `list`, `table`, or `textarea`.
- Avoid negative dimensions; clamp to zero or a minimum.
