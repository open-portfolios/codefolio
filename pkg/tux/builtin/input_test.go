package builtin

import (
	"testing"

	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/debug"
)

func TestInputRenderWritesContent(t *testing.T) {
	renderer := debug.NewRenderer(1, 5)
	component := Input(InputProps{Width: 5, Content: "abc"})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.String(0, 0, 5), "abc  "; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestInputRenderUsesPosition(t *testing.T) {
	renderer := debug.NewRenderer(3, 5)
	component := Input(InputProps{Row: 1, Column: 2, Width: 2, Content: "xy"})

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
	component := Input(InputProps{Width: 2, Content: "abc"})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.String(0, 0, 2), "ab"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestInputRenderWritesColors(t *testing.T) {
	renderer := debug.NewRenderer(1, 1)
	component := Input(InputProps{
		Width:      1,
		Content:    "x",
		Foreground: tux.ColorRed,
		Background: tux.ColorBlue,
	})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.Buf[0][0], (tux.Cell{Ch: 'x', Foreground: tux.ColorRed, Background: tux.ColorBlue}); got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestInputRenderQueuesCursorWhenFocused(t *testing.T) {
	renderer := debug.NewRenderer(2, 5)
	component := Input(InputProps{Row: 1, Column: 2, Width: 3, Cursor: 1, Focused: true})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	row, column, queued := renderer.CursorPos()
	if !queued {
		t.Fatal("cursor move was not queued")
	}
	if row != 1 || column != 3 {
		t.Fatalf("got (%d,%d), want (1,3)", row, column)
	}
}

func TestInputRenderDoesNotQueueCursorWhenUnfocused(t *testing.T) {
	renderer := debug.NewRenderer(1, 3)
	component := Input(InputProps{Width: 3, Cursor: 1})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if _, _, queued := renderer.CursorPos(); queued {
		t.Fatal("cursor move was queued")
	}
}

func TestInputRenderClampsCursor(t *testing.T) {
	renderer := debug.NewRenderer(1, 3)
	component := Input(InputProps{Width: 3, Cursor: 9, Focused: true})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	row, column, queued := renderer.CursorPos()
	if !queued {
		t.Fatal("cursor move was not queued")
	}
	if row != 0 || column != 2 {
		t.Fatalf("got (%d,%d), want (0,2)", row, column)
	}
}
