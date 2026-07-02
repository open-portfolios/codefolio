package builtin

import (
	"testing"

	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/debug"
)

func TestBoxRenderSkipsEmptySize(t *testing.T) {
	renderer := debug.NewRenderer(1, 1)
	component := Box(BoxProps{Width: 0, Height: 1, Content: "x"})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got := renderer.Buf[0][0]; got != (tux.Cell{}) {
		t.Fatalf("got %#v, want zero cell", got)
	}
}

func TestBoxRenderFillsRectangle(t *testing.T) {
	renderer := debug.NewRenderer(2, 3)
	component := Box(BoxProps{Width: 3, Height: 2})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	for row := range 2 {
		for column := range 3 {
			if got, want := renderer.Buf[row][column].Ch, rune(' '); got != want {
				t.Fatalf("cell (%d,%d) got %q, want %q", row, column, got, want)
			}
		}
	}
}

func TestBoxRenderUsesPosition(t *testing.T) {
	renderer := debug.NewRenderer(3, 4)
	component := Box(BoxProps{Row: 1, Column: 2, Width: 2, Height: 1, Content: "xy"})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.Buf[1][2].Ch, rune('x'); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := renderer.Buf[1][3].Ch, rune('y'); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := renderer.Buf[0][0]; got != (tux.Cell{}) {
		t.Fatalf("got %#v, want zero cell", got)
	}
}

func TestBoxRenderWritesContentRowMajor(t *testing.T) {
	renderer := debug.NewRenderer(2, 3)
	component := Box(BoxProps{Width: 3, Height: 2, Content: "abcdef"})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.String(0, 0, 3)+renderer.String(1, 0, 3), "abcdef"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBoxRenderWritesWideRunes(t *testing.T) {
	renderer := debug.NewRenderer(1, 4)
	component := Box(BoxProps{Width: 4, Height: 1, Content: "你a"})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.String(0, 0, 4), "你a "; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := renderer.Buf[0][0].Width, uint8(2); got != want {
		t.Fatalf("got width %d, want %d", got, want)
	}
	if got, want := renderer.Buf[0][1].Width, uint8(0); got != want {
		t.Fatalf("got continuation width %d, want %d", got, want)
	}
}

func TestBoxRenderTruncatesContent(t *testing.T) {
	renderer := debug.NewRenderer(1, 2)
	component := Box(BoxProps{Width: 2, Height: 1, Content: "abc"})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.String(0, 0, 2), "ab"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBoxRenderWritesColors(t *testing.T) {
	renderer := debug.NewRenderer(1, 1)
	component := Box(BoxProps{
		Width:      1,
		Height:     1,
		Content:    "x",
		Foreground: tux.ColorRed,
		Background: tux.ColorBlue,
	})

	if err := component.Render(tux.BuildContext{}, renderer); err != nil {
		t.Fatal(err)
	}

	if got, want := renderer.Buf[0][0], (tux.Cell{Ch: 'x', Width: 1, Foreground: tux.ColorRed, Background: tux.ColorBlue}); got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
