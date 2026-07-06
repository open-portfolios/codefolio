# Huh Forms Reference

Huh is a high-level form and prompt framework built on Bubble Tea, Bubbles, and Lip Gloss. Use it when the app primarily needs to collect structured input rather than manually composing individual inputs.

## Core Concepts

- `Form`: top-level model; can run standalone or be embedded in Bubble Tea.
- `Group`: a page/section containing fields.
- `Field`: one input element.
- `Theme`: Lip Gloss-based visual style.

## Field Types

| Field                     | Use                                  |
| ------------------------- | ------------------------------------ |
| `huh.NewInput()`          | Single-line string input.            |
| `huh.NewText()`           | Multi-line text input.               |
| `huh.NewSelect[T]()`      | Single choice from typed options.    |
| `huh.NewMultiSelect[T]()` | Multiple choices from typed options. |
| `huh.NewConfirm()`        | Yes/no boolean.                      |
| `huh.NewFilePicker()`     | File or directory selection.         |
| `huh.NewNote()`           | Informational note.                  |

## Standalone Form

```go
var name string
var confirm bool

form := huh.NewForm(
	huh.NewGroup(
		huh.NewInput().
			Title("Name").
			Placeholder("Ada Lovelace").
			Value(&name).
			Validate(huh.ValidateNotEmpty()),
		huh.NewConfirm().
			Title("Continue?").
			Value(&confirm),
	),
)

if err := form.Run(); err != nil {
	return err
}
```

## Select Options

```go
var flavor string

huh.NewSelect[string]().
	Title("Choose flavor").
	Options(
		huh.NewOption("Vanilla", "vanilla"),
		huh.NewOption("Chocolate", "chocolate"),
	).
	Value(&flavor)
```

## Embedding in Bubble Tea

Use this when the form is one screen inside a larger app:

```go
type model struct {
	form *huh.Form
}

func (m model) Init() tea.Cmd {
	return m.form.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.form.Update(msg)
	if form, ok := updated.(*huh.Form); ok {
		m.form = form
	}
	return m, cmd
}

func (m model) View() tea.View {
	if m.form.State == huh.StateCompleted {
		return tea.NewView("Done\n")
	}
	return tea.NewView(m.form.View())
}
```

## Use Huh When

- You need a wizard, setup flow, configuration prompt, survey, or guided input.
- Fields need validation, themes, conditional visibility, or accessibility support.
- You do not need highly custom per-frame layout or widget composition.

Use raw Bubble Tea + Bubbles when the UI is an application screen with many interacting panes, custom navigation, or dynamic data display.
