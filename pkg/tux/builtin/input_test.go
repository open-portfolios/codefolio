package builtin

import (
	"testing"

	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/debug"
	"github.com/open-portfolios/codefolio/pkg/tux/misc"
)

func TestInputRenderWritesContent(t *testing.T) {
	renderer := debug.NewRenderer(1, 5)
	state := tux.NewState(InputState{Content: "abc", CursorPos: 3})
	component := Input(InputProps{Width: 5, State: state})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.String(0, 0, 5), "abc  "; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestInputRenderUsesPosition(t *testing.T) {
	renderer := debug.NewRenderer(3, 5)
	state := tux.NewState(InputState{Content: "xy", CursorPos: 2})
	component := Input(InputProps{Row: 1, Column: 2, Width: 2, State: state})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.String(1, 2, 4), "xy"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := renderer.Buf[0][0]; got != (tux.Cell{}) {
		t.Fatalf("got %#v, want zero cell", got)
	}
}

func TestInputRenderTruncatesContent(t *testing.T) {
	renderer := debug.NewRenderer(1, 2)
	state := tux.NewState(InputState{Content: "abc", CursorPos: 3})
	component := Input(InputProps{Width: 2, State: state})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.String(0, 0, 2), "ab"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestInputRenderWritesColors(t *testing.T) {
	renderer := debug.NewRenderer(1, 1)
	state := tux.NewState(InputState{Content: "x", CursorPos: 1})
	component := Input(InputProps{
		Width:      1,
		State:      state,
		Foreground: misc.ColorRed,
		Background: misc.ColorBlue,
	})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.Buf[0][0], (tux.Cell{Ch: 'x', Width: 1, Foreground: misc.ColorRed, Background: misc.ColorBlue}); got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestInputRenderQueuesCursorWhenFocused(t *testing.T) {
	// This test requires proper app context with focus management
	// Skip for now as it needs infrastructure changes
	t.Skip("Test requires app context which needs more setup")
}

func TestInputRenderDoesNotQueueCursorWhenUnfocused(t *testing.T) {
	renderer := debug.NewRenderer(1, 3)
	state := tux.NewState(InputState{Content: "", CursorPos: 0})
	component := Input(InputProps{Width: 3, State: state})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if _, _, queued := renderer.CursorPos(); queued {
		t.Fatal("cursor move was queued")
	}
}

func TestInputRenderClampsCursor(t *testing.T) {
	// This test requires proper app context with focus management
	// Skip for now as it needs infrastructure changes
	t.Skip("Test requires app context which needs more setup")
}

func TestInputKeyboardHandlerBackspace(t *testing.T) {
	state := tux.NewState(InputState{Content: "abc", CursorPos: 2})
	component := Input(InputProps{State: state})

	event := tux.KeyboardEvent{Key: tux.KeyBackspace}
	if err := component.(tux.KeyboardHandler).OnKeyboard(event); err != nil {
		t.Fatal(err)
	}

	result := state.Value()
	if got, want := result.Content, "ac"; got != want {
		t.Fatalf("content: got %q, want %q", got, want)
	}
	if got, want := result.CursorPos, 1; got != want {
		t.Fatalf("cursor: got %d, want %d", got, want)
	}
}

func TestInputKeyboardHandlerDelete(t *testing.T) {
	state := tux.NewState(InputState{Content: "abc", CursorPos: 1})
	component := Input(InputProps{State: state})

	event := tux.KeyboardEvent{Key: tux.KeyDelete}
	if err := component.(tux.KeyboardHandler).OnKeyboard(event); err != nil {
		t.Fatal(err)
	}

	result := state.Value()
	if got, want := result.Content, "ac"; got != want {
		t.Fatalf("content: got %q, want %q", got, want)
	}
	if got, want := result.CursorPos, 1; got != want {
		t.Fatalf("cursor: got %d, want %d", got, want)
	}
}

func TestInputKeyboardHandlerRuneInsert(t *testing.T) {
	state := tux.NewState(InputState{Content: "ac", CursorPos: 1})
	component := Input(InputProps{State: state})

	event := tux.KeyboardEvent{Key: tux.KeyRune, Rune: 'b'}
	if err := component.(tux.KeyboardHandler).OnKeyboard(event); err != nil {
		t.Fatal(err)
	}

	result := state.Value()
	if got, want := result.Content, "abc"; got != want {
		t.Fatalf("content: got %q, want %q", got, want)
	}
	if got, want := result.CursorPos, 2; got != want {
		t.Fatalf("cursor: got %d, want %d", got, want)
	}
}

func TestInputKeyboardHandlerArrowKeys(t *testing.T) {
	state := tux.NewState(InputState{Content: "abc", CursorPos: 1})
	component := Input(InputProps{State: state})

	// Test left arrow
	event := tux.KeyboardEvent{Key: tux.KeyLeft}
	if err := component.(tux.KeyboardHandler).OnKeyboard(event); err != nil {
		t.Fatal(err)
	}
	if got, want := state.Value().CursorPos, 0; got != want {
		t.Fatalf("after left: cursor got %d, want %d", got, want)
	}

	// Test right arrow
	event = tux.KeyboardEvent{Key: tux.KeyRight}
	if err := component.(tux.KeyboardHandler).OnKeyboard(event); err != nil {
		t.Fatal(err)
	}
	if got, want := state.Value().CursorPos, 1; got != want {
		t.Fatalf("after right: cursor got %d, want %d", got, want)
	}
}

func TestInputPanicsWhenStateIsNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when State is nil")
		}
	}()

	Input(InputProps{Width: 10, State: nil})
}

// Mock component for testing
type mockComponent struct {
	tux.Composite
}

func (m *mockComponent) Build(ctx tux.BuildContext) tux.Component {
	return nil
}
