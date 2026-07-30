package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cylixlee/tux/app"
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
	"github.com/cylixlee/tux/terminal"
	"github.com/open-portfolios/codefolio/cmd/cli/components"
	"github.com/open-portfolios/codefolio/cmd/cli/controller"
	"github.com/open-portfolios/codefolio/internal/conf"
)

func TestAppRendersOpenCodeSessionSurfaces(t *testing.T) {
	mockTerminal := terminal.NewMock(130, 30)
	root := &App{
		controller:   controller.New(&conf.Struct{Model: "test-model"}, nil, nil, nil, nil),
		workDir:      "~/workspace",
		editor:       state.New(builtin.TextareaState{PreferredColumn: -1}),
		viewport:     state.New(builtin.ViewportState{FollowEnd: true}),
		spinner:      state.New(0),
		askOpen:      state.New(false),
		approvalOpen: state.New(false),
	}
	root.handleEditorKey = root.editorKey

	runtime := app.New(app.AppConfig{
		Root:       root,
		Terminal:   mockTerminal,
		Background: components.Theme.Background,
	})
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- runtime.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("app run returned %v", err)
			}
		case <-time.After(time.Second):
			t.Error("app did not stop")
		}
	})

	deadline := time.Now().Add(time.Second)
	for mockTerminal.Output() == "" && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	for _, point := range []struct {
		x, y int
		bg   style.Color
	}{
		{x: 0, y: 0, bg: components.Theme.Background},
		{x: 129, y: 0, bg: components.Theme.BackgroundPanel},
		{x: 129, y: 29, bg: components.Theme.BackgroundPanel},
		{x: 90, y: 3, bg: components.Theme.BackgroundPanel},
		{x: 90, y: 4, bg: components.Theme.BackgroundPanel},
		{x: 2, y: 22, bg: components.Theme.BackgroundElement},
		{x: 2, y: 23, bg: components.Theme.BackgroundElement},
		{x: 2, y: 24, bg: components.Theme.BackgroundElement},
		{x: 2, y: 25, bg: components.Theme.BackgroundElement},
		{x: 2, y: 26, bg: components.Theme.BackgroundElement},
		{x: 2, y: 27, bg: components.Theme.Background},
		{x: 2, y: 28, bg: components.Theme.Background},
		{x: 2, y: 29, bg: components.Theme.Background},
		{x: 2, y: 29, bg: components.Theme.Background},
	} {
		if cell := mockTerminal.Cell(point.x, point.y); cell.Style.Bg != point.bg {
			t.Fatalf("screen cell (%d, %d) background = %#v, want %#v", point.x, point.y, cell.Style.Bg, point.bg)
		}
	}
	for _, expected := range []string{"Plan · test-model", "Context", "MCP", "No servers configured"} {
		found := false
		for y := range 30 {
			if strings.Contains(mockTerminal.Line(y), expected) {
				found = true
				break
			}
		}
		if !found {
			lines := make([]string, 30)
			for y := range lines {
				lines[y] = mockTerminal.Line(y)
			}
			t.Fatalf("mock terminal frame does not contain %q:\n%s", expected, strings.Join(lines, "\n"))
		}
	}
	for _, unexpected := range []string{"Todo", "Inspect the workspace", "Plan the next change"} {
		for y := range 30 {
			if strings.Contains(mockTerminal.Line(y), unexpected) {
				t.Fatalf("mock terminal frame must not contain removed sidebar text %q", unexpected)
			}
		}
	}
	for y := range 30 {
		if index := strings.Index(mockTerminal.Line(y), "Untitled session"); index >= 0 {
			if index != 90 {
				t.Fatalf("sidebar title starts at column %d, want panel start plus two columns (90)", index)
			}
			break
		}
	}
	footerRow := -1
	for y := range 30 {
		if strings.Contains(mockTerminal.Line(y), "● Codefolio") {
			footerRow = y
			break
		}
	}
	if footerRow < 0 || footerRow+1 >= 30 {
		t.Fatalf("sidebar footer is missing or has no bottom row")
	}
	if strings.TrimSpace(mockTerminal.Line(footerRow + 1)[88:]) != "" {
		t.Fatalf("sidebar footer must leave exactly one panel row below it:\n%s\n%s", mockTerminal.Line(footerRow), mockTerminal.Line(footerRow+1))
	}
	output := mockTerminal.Output()
	if strings.Contains(output, "\x1b[49m") {
		t.Fatalf("render output resets to the terminal default background:\n%q", output)
	}
}
