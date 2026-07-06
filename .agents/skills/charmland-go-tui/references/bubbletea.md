# Bubble Tea v2 Reference

Bubble Tea is the core runtime for interactive Go TUIs. It follows the Elm Architecture: state lives in a model, messages describe events, updates transform state, and views render state.

## Core Types

```go
type Model interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() tea.View
}
```

- `tea.Msg` is any event value: key press, resize, timer tick, I/O result, or custom message.
- `tea.Cmd` is `func() tea.Msg`; use it for I/O and async work.
- `tea.View` is the rendered terminal view. Use `tea.NewView(content)` for ordinary string content.
- `tea.Program` is the runtime created with `tea.NewProgram(model, opts...)` and run with `p.Run()`.

## Basic Program

```go
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	cursor int
	items  []string
}

func initialModel() model {
	return model{items: []string{"Alpha", "Beta", "Gamma"}}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	s := "Choose an item:\n\n"
	for i, item := range m.items {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, item)
	}
	return tea.NewView(s)
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

## Commands and Messages

Use custom message types for async results:

```go
type loadedMsg struct{ value string }
type errMsg struct{ err error }

func loadData() tea.Msg {
	value, err := expensiveCall()
	if err != nil {
		return errMsg{err: err}
	}
	return loadedMsg{value: value}
}

func (m model) Init() tea.Cmd {
	return loadData
}
```

Use `tea.Batch(cmds...)` for independent concurrent work and `tea.Sequence(cmds...)` when order matters.

## Responsive Layout

Handle `tea.WindowSizeMsg` and store dimensions in the model. Recalculate child component dimensions when the terminal changes size.

```go
case tea.WindowSizeMsg:
	m.width = msg.Width
	m.height = msg.Height
	m.viewport.SetWidth(msg.Width)
	m.viewport.SetHeight(msg.Height - headerHeight - footerHeight)
```

## View Features

Most apps return `tea.NewView(content)`. For terminal features, set fields on the returned view:

```go
func (m model) View() tea.View {
	v := tea.NewView(m.content)
	v.AltScreen = true
	v.WindowTitle = "My TUI"
	return v
}
```

Use this declarative v2 style instead of older imperative screen commands when possible.

## Testing Strategy

- Test pure helper functions directly.
- Test `Update` by constructing a model, sending messages, and asserting the returned model state.
- Avoid snapshotting full ANSI output unless the view is stable and small.
- Keep I/O behind commands so it can be tested separately.
