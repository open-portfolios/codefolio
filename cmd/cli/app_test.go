package main

import (
	"testing"

	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/state"
	"github.com/open-portfolios/codefolio/cmd/cli/controller"
)

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
