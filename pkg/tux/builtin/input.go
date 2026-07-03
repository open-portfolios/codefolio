package builtin

import (
	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/misc"
)

var (
	_ tux.Component       = (*input)(nil)
	_ tux.KeyboardHandler = (*input)(nil)
)

// InputState represents the state of an Input component.
type InputState struct {
	Content   string
	CursorPos int
}

// NewInputState creates an InputState with the cursor at the end of the content.
func NewInputState(content string) InputState {
	return InputState{
		Content:   content,
		CursorPos: len([]rune(content)),
	}
}

type input struct {
	*box

	state        *tux.State[InputState]
	autoFocus    bool
	focusApplied bool
	onChange     func(InputState) error
}

type InputProps struct {
	Row        int
	Column     int
	Width      int
	State      *tux.State[InputState] // Required
	AutoFocus  bool
	Foreground misc.Color
	Background misc.Color
	OnChange   func(InputState) error
}

func Input(props InputProps, children ...tux.Component) tux.Component {
	if props.State == nil {
		panic("tux: Input.State must not be nil")
	}

	return &input{
		box: &box{
			row:        props.Row,
			column:     props.Column,
			width:      props.Width,
			height:     1,
			content:    "",
			foreground: props.Foreground,
			background: props.Background,
		},
		state:     props.State,
		autoFocus: props.AutoFocus,
		onChange:  props.OnChange,
	}
}

func (i *input) OnKeyboard(e tux.KeyboardEvent) error {
	// Ignore modified input
	if e.Mod != 0 {
		return nil
	}

	current := i.state.Value()
	content := current.Content
	cursor := current.CursorPos

	newContent := content
	newCursor := cursor
	changed := false

	switch e.Key {
	case tux.KeyBackspace:
		if cursor > 0 {
			runes := []rune(content)
			newContent = string(append(runes[:cursor-1], runes[cursor:]...))
			newCursor = cursor - 1
			changed = true
		}

	case tux.KeyDelete:
		runes := []rune(content)
		if cursor < len(runes) {
			newContent = string(append(runes[:cursor], runes[cursor+1:]...))
			changed = true
		}

	case tux.KeyLeft:
		if cursor > 0 {
			newCursor = cursor - 1
			changed = true
		}

	case tux.KeyRight:
		if cursor < len([]rune(content)) {
			newCursor = cursor + 1
			changed = true
		}

	case tux.KeyRune:
		if e.Rune >= 32 { // Printable character
			runes := []rune(content)
			result := make([]rune, 0, len(runes)+1)
			result = append(result, runes[:cursor]...)
			result = append(result, e.Rune)
			result = append(result, runes[cursor:]...)
			newContent = string(result)
			newCursor = cursor + 1
			changed = true
		}
	}

	if changed {
		newState := InputState{
			Content:   newContent,
			CursorPos: newCursor,
		}
		i.state.Set(newState)

		if i.onChange != nil {
			return i.onChange(newState)
		}
	}

	return nil
}

func (i *input) Render(build tux.BuildContext, ctx tux.RenderContext) error {
	// AutoFocus logic: only the first AutoFocus component gets focus
	if i.autoFocus && !i.focusApplied {
		if app := build.App(); app != nil {
			if !app.IsFocusApplied() {
				app.SetFocus(i)
				app.MarkFocusApplied()
				i.focusApplied = true
			}
		}
	}

	if i.box.width <= 0 {
		return nil
	}

	// Read current state
	inputState := i.state.Get(build)
	content := inputState.Content
	cursor := inputState.CursorPos

	// Ensure cursor is within bounds
	runeLen := len([]rune(content))
	if cursor > runeLen {
		cursor = runeLen
	}
	if cursor < 0 {
		cursor = 0
	}

	// Update box content
	i.box.content = content

	// Render content
	if err := i.box.Render(build, ctx); err != nil {
		return err
	}

	// If this is the focused component, set cursor position
	if app := build.App(); app != nil && app.GetFocus() == i {
		// Clamp to input box width (allow cursor at right edge)
		cursorVisualPos := min(misc.StringWidth(misc.PrefixRunes(content, cursor)), i.box.width)
		ctx.QueueCursorMove(i.box.row, i.box.column+cursorVisualPos)
	}

	return nil
}
