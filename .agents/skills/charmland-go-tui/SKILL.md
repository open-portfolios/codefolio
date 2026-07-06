---
name: charmland-go-tui
description: "Build Go terminal user interfaces with the charm.land ecosystem: Bubble Tea v2, Lip Gloss v2, Bubbles v2, Huh, Glamour, and charmbracelet/log. Use when creating, debugging, refactoring, or explaining Go TUI applications, terminal forms, terminal components, Markdown rendering, styling/layout, or logging in TUI programs."
license: MIT
---

# Charmland Go TUI

Use this skill for Go terminal UI work based on Charm's `charm.land` ecosystem. Focus on the core TUI stack:

- `charm.land/bubbletea/v2` for the Elm-style TUI runtime.
- `charm.land/lipgloss/v2` for styling, layout, tables, lists, and trees.
- `charm.land/bubbles/v2` for reusable interactive components.
- `charm.land/huh/v2` for higher-level interactive forms and prompts.
- `charm.land/glamour/v2` for Markdown rendering in the terminal.
- `charm.land/log/v2` for structured logging that fits terminal apps.

Do not introduce `harmonica` or `wish` by default. They are optional adjacent tools: use animation libraries only when the user explicitly needs physics animation, and use SSH-serving libraries only when the user explicitly wants to serve a TUI over SSH.

## When to Use This Skill

- The user asks to build a Go TUI, CLI UI, terminal dashboard, terminal wizard, text UI, or interactive terminal app.
- The user mentions Bubble Tea, Lip Gloss, Bubbles, Huh, Glamour, or Charmbracelet libraries.
- The task involves terminal layout, styles, key handling, forms, lists, tables, text inputs, viewport scrolling, Markdown display, progress/spinner UI, or TUI logging.
- The task is debugging or upgrading a Charm TUI app, especially v1 to v2 API changes.

## Default Workflow

1. Inspect `go.mod` first to confirm module paths and versions.
2. Prefer existing project patterns over new abstractions.
3. Use Bubble Tea for app state and event flow; use Lip Gloss only for rendering and layout.
4. Use Bubbles for interactive widgets instead of rebuilding text inputs, lists, tables, spinners, viewports, or paginators.
5. Use Huh when the main task is collecting structured user input through prompts/forms.
6. Use Glamour when terminal output includes Markdown, docs, help pages, changelogs, or rich prose.
7. Use `charm.land/log/v2` for debugging/logging; avoid printing to stdout while Bubble Tea owns the terminal.
8. Run `gofmt` and the relevant tests or build command after edits.

## Version Rules

This workspace uses the Charm v2 import paths for the core stack:

```go
tea "charm.land/bubbletea/v2"
"charm.land/lipgloss/v2"
"charm.land/bubbles/v2/textinput"
"charm.land/bubbles/v2/viewport"
"charm.land/huh/v2"
"charm.land/glamour/v2"
"charm.land/log/v2"
```

Important Bubble Tea v2 rule: application models implement `View() tea.View`, not `View() string`. Most Bubbles components still render `View() string`; compose those strings in the parent model and return `tea.NewView(content)`.

```go
func (m model) View() tea.View {
	content := m.input.View() + "\n" + m.viewport.View()
	return tea.NewView(content)
}
```

## Library Selection

| Need                                                     | Use          | Notes                                                             |
| -------------------------------------------------------- | ------------ | ----------------------------------------------------------------- |
| App runtime, key handling, messages, async commands      | `Bubble Tea` | Model owns state; Update handles events; View renders from state. |
| Colors, borders, padding, alignment, layout              | `Lip Gloss`  | Keep styling in small reusable `lipgloss.Style` values.           |
| Text inputs, tables, lists, spinners, progress, viewport | `Bubbles`    | Compose component models inside your app model.                   |
| Structured prompts and multi-step forms                  | `Huh`        | Faster than hand-building forms from individual Bubbles.          |
| Markdown output                                          | `Glamour`    | Render Markdown, then style/place with Lip Gloss if needed.       |
| Debug and structured logging                             | `Log`        | Prefer file/stderr logging; do not corrupt the TUI renderer.      |

## Bubble Tea Architecture Rules

- Keep all mutable UI state in the model.
- Only mutate state inside `Update`; keep `View` deterministic and side-effect free.
- Represent external results as custom `tea.Msg` types.
- Return `tea.Cmd` for I/O, timers, subprocesses, and async work; never block in `Update`.
- Use `tea.Batch` for independent concurrent commands and `tea.Sequence` for ordered commands.
- Always handle quit keys like `ctrl+c` and usually `q` where appropriate.
- Handle `tea.WindowSizeMsg` for responsive layouts.

Minimal shape:

```go
type model struct {
	ready  bool
	width  int
	height int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
	}
	return m, nil
}

func (m model) View() tea.View {
	if !m.ready {
		return tea.NewView("initializing...\n")
	}
	return tea.NewView("Hello from Bubble Tea\n")
}
```

## Component Composition

When embedding Bubbles components:

- Store component models as fields on the parent model.
- Call the component's `Init` command from parent `Init` when it needs one.
- In parent `Update`, delegate messages to focused or relevant child components and store the returned child model.
- In parent `View`, call child `View()` methods and compose their strings with Lip Gloss.

Use references for specifics:

- [Bubble Tea architecture](references/bubbletea.md)
- [Lip Gloss styling and layout](references/lipgloss.md)
- [Bubbles components](references/bubbles.md)
- [Huh forms](references/huh.md)
- [Glamour Markdown rendering](references/glamour.md)
- [Logging](references/logging.md)
- [Patterns and pitfalls](references/patterns.md)

## Quality Bar

- Prefer the smallest correct implementation.
- Do not hand-roll terminal escape codes unless the library lacks the feature.
- Keep style definitions separate from business logic when the view grows.
- Avoid global state except package-level immutable styles.
- Add tests for pure state transitions, parsing, filtering, validation, and formatting logic.
- For visual behavior, prefer small deterministic helpers that can be unit-tested outside the terminal runtime.
