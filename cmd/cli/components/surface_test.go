package components

import (
	"strings"
	"testing"

	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
)

func TestSidebarKeepsPanelBackgroundBehindText(t *testing.T) {
	ctx := renderer.Context{SizeFn: func() (int, int) { return 42, 12 }}
	sidebar := NewSidebar(ctx, SidebarProps{WorkDir: "~/workspace"})
	buffer := surfaceBuffer(42, 12)
	sidebar.Render(ctx).Draw(buffer, 0, 0, 42, 12)

	for i, cell := range buffer.Cells {
		if cell.Style.Bg != Theme.BackgroundPanel {
			t.Fatalf("sidebar cell %d background = %#v, want panel %#v", i, cell.Style.Bg, Theme.BackgroundPanel)
		}
	}
}

func TestPromptFillsComposerSurfaceAndDoesNotResetBackground(t *testing.T) {
	ctx := renderer.Context{SizeFn: func() (int, int) { return 60, 9 }}
	prompt := NewPrompt(ctx, PromptProps{
		State:   state.New(builtin.TextareaState{}),
		Model:   "test-model",
		Profile: "build",
		WorkDir: "~/workspace",
	})
	buffer := surfaceBuffer(60, 9)
	prompt.Render(ctx).Draw(buffer, 0, 0, 60, 9)

	for x := range 60 {
		cell := buffer.Cells[x]
		if cell.Style.Bg != Theme.BackgroundElement {
			t.Fatalf("composer cell (%d, 0) background = %#v, want element %#v", x, cell.Style.Bg, Theme.BackgroundElement)
		}
	}
	for _, row := range []int{0, 1, 2, 3, 4} {
		cell := buffer.Cells[row*buffer.Width]
		if cell.Rune != '┃' || cell.Style.Fg != Theme.Secondary {
			t.Fatalf("composer rail cell (0, %d) = %#v, want colored rail", row, cell)
		}
	}
	if height := prompt.Render(ctx).LayoutHeight(); height != 8 {
		t.Fatalf("prompt height = %d, want composer padding, status, and bottom space", height)
	}

	for _, cell := range buffer.Cells {
		if cell.Rune != 0 && cell.Style.Bg == style.ColorDefault {
			t.Fatalf("visible prompt content reset to terminal default background: %#v", cell)
		}
	}

	ansi := strings.Join(buffer.Diff(renderer.NewCellBuffer(60, 9)), "")
	if strings.Contains(ansi, "\x1b[49m") {
		t.Fatalf("prompt output contains a default-background reset: %q", ansi)
	}
}

func TestPromptPlanRailUsesPrimaryColor(t *testing.T) {
	ctx := renderer.Context{SizeFn: func() (int, int) { return 60, 9 }}
	prompt := NewPrompt(ctx, PromptProps{State: state.New(builtin.TextareaState{}), Profile: "plan"})
	buffer := surfaceBuffer(60, 9)
	prompt.Render(ctx).Draw(buffer, 0, 0, 60, 9)
	if cell := buffer.Cells[0]; cell.Rune != '┃' || cell.Style.Fg != Theme.Primary {
		t.Fatalf("plan composer rail = %#v, want primary rail", cell)
	}
}

func TestPromptGrowsForMultilineInputUntilItsHeightLimit(t *testing.T) {
	ctx := renderer.Context{SizeFn: func() (int, int) { return 60, 20 }}
	prompt := NewPrompt(ctx, PromptProps{
		State:   state.New(builtin.TextareaState{Value: "one\ntwo\nthree"}),
		Model:   "test-model",
		WorkDir: "~/workspace",
	})
	if got := prompt.Render(ctx).LayoutHeight(); got != 10 {
		t.Fatalf("multiline prompt height = %d, want 10", got)
	}

	long := NewPrompt(ctx, PromptProps{
		State:   state.New(builtin.TextareaState{Value: "1\n2\n3\n4\n5\n6\n7\n8\n9\n10"}),
		Model:   "test-model",
		WorkDir: "~/workspace",
	})
	if got := long.Render(ctx).LayoutHeight(); got != 15 {
		t.Fatalf("bounded prompt height = %d, want 15", got)
	}
}

func surfaceBuffer(width, height int) *renderer.CellBuffer {
	buffer := renderer.NewCellBuffer(width, height)
	buffer.Fill(0, 0, width, height, renderer.Cell{Rune: ' ', Style: style.Style{Bg: Theme.Background}})
	return buffer
}
