package main

import (
	"testing"

	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/state"
	"github.com/open-portfolios/codefolio/cmd/cli/controller"
	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/svc"
)

func TestExitCommandUsesAttachedRuntimeStop(t *testing.T) {
	for _, command := range []string{"exit", "quit"} {
		called := false
		root := &App{controller: &controller.Controller{}, quit: func() { called = true }}
		root.executeLocal(command, "")
		if !called {
			t.Fatalf("/%s must stop the attached TUX runtime", command)
		}
	}
}

func TestEditorCtrlCClearsOnlyNonEmptyInput(t *testing.T) {
	root := &App{editor: state.New(builtin.TextareaState{Value: "draft", Cursor: 5, PreferredColumn: -1}), controller: &controller.Controller{}}
	ctrlC := input.KeyEvent{Rune: 'c', Modifiers: input.ModCtrl}
	if !root.editorKey(ctrlC) {
		t.Fatal("non-empty editor should consume Ctrl+C to clear")
	}
	if got := root.editor.Value().Value; got != "" {
		t.Fatalf("editor value = %q, want cleared", got)
	}
	if root.editorKey(ctrlC) {
		t.Fatal("empty editor must leave Ctrl+C to the TUX app fallback")
	}
}

func TestComposerFocusRestoresOnlyWhenTimelineFollowsEnd(t *testing.T) {
	root := &App{
		controller:      &controller.Controller{},
		viewport:        state.New(builtin.ViewportState{FollowEnd: true}),
		composerEnabled: false,
	}
	if root.ComposerDisabled() {
		t.Fatal("idle composer should be enabled")
	}
	if !root.FocusComposer() {
		t.Fatal("composer should focus when it becomes enabled at the timeline end")
	}

	root.composerEnabled = false
	root.viewport.Set(builtin.ViewportState{FollowEnd: false})
	if root.ComposerDisabled() {
		t.Fatal("idle composer should remain enabled")
	}
	if root.FocusComposer() {
		t.Fatal("composer must not reclaim focus when the timeline is scrolled away from the end")
	}
}

func TestComposerTabSwitchesProfiles(t *testing.T) {
	agent := &svc.Agent{Mode: svc.ModePlan}
	root := &App{editor: state.New(builtin.TextareaState{PreferredColumn: -1}), controller: controller.New(&conf.Struct{}, agent, nil, nil, nil)}
	if !root.editorKey(input.KeyEvent{Key: input.KeyTab}) || agent.Mode != svc.ModeDefault {
		t.Fatalf("Tab should select build mode, got %v", agent.Mode)
	}
	if !root.editorKey(input.KeyEvent{Key: input.KeyTab}) || agent.Mode != svc.ModePlan {
		t.Fatalf("second Tab should select plan mode, got %v", agent.Mode)
	}
	if root.editor.Value().Value != "" {
		t.Fatalf("profile switching must preserve composer text, got %q", root.editor.Value().Value)
	}
}
