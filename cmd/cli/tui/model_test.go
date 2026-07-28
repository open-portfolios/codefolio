package tui

import (
	"testing"

	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/state"

	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
)

func TestWrapVisualPreservesContentAndWidth(t *testing.T) {
	lines := wrapVisual("hello world", 5, primary, 0, "", 0, "")
	if len(lines) != 3 {
		t.Fatalf("line count = %d, want 3", len(lines))
	}
	if lines[0].text != "hello" || lines[1].text != " worl" || lines[2].text != "d" {
		t.Fatalf("wrapped lines = %#v", lines)
	}
}

func TestToolLabelHandlesKnownAndUnknownTools(t *testing.T) {
	if got := toolLabel(&toolView{Name: "Bash", Input: `{"command":"go test ./..."}`}); got != "Bash go test ./..." {
		t.Fatalf("Bash label = %q", got)
	}
	if got := toolLabel(&toolView{Name: "CustomTool"}); got != "CustomTool" {
		t.Fatalf("unknown label = %q", got)
	}
}

func TestEditorCtrlCClearsOnlyNonEmptyInput(t *testing.T) {
	m := &Model{editor: state.New(builtin.TextareaState{Value: "draft", Cursor: 5, PreferredColumn: -1})}
	ctrlC := input.KeyEvent{Rune: 'c', Modifiers: input.ModCtrl}
	if !m.handleEditorKey(ctrlC) {
		t.Fatal("non-empty editor should consume Ctrl+C to clear")
	}
	if got := m.editor.Value().Value; got != "" {
		t.Fatalf("editor value = %q, want cleared", got)
	}
	if m.handleEditorKey(ctrlC) {
		t.Fatal("empty editor must leave Ctrl+C to the TUX app fallback")
	}
}

func TestHistoryRestoresDraft(t *testing.T) {
	m := &Model{
		editor:    state.New(builtin.TextareaState{Value: "draft", Cursor: 5, PreferredColumn: -1}),
		history:   []string{"first", "second"},
		historyAt: 2,
	}
	m.historyUp()
	if got := m.editor.Value().Value; got != "second" {
		t.Fatalf("history up = %q, want second", got)
	}
	m.historyDown()
	if got := m.editor.Value().Value; got != "draft" {
		t.Fatalf("history down = %q, want restored draft", got)
	}
}

func TestMarkdownVisualRendersCommonAgentBlocks(t *testing.T) {
	lines := markdownVisual("# Plan\n> inspect first\n- implement\n```go\nfmt.Println(\"ok\")\n```", 40, 0)
	if len(lines) != 5 {
		t.Fatalf("line count = %d, want 5", len(lines))
	}
	if lines[0].text != "Plan" || lines[0].attrs != 1 {
		t.Fatalf("heading = %#v", lines[0])
	}
	if lines[1].text != "| inspect first" {
		t.Fatalf("quote = %#v", lines[1])
	}
	if lines[2].text != "• implement" {
		t.Fatalf("list = %#v", lines[2])
	}
	if lines[3].text != "  go" || lines[3].bg != codeBg {
		t.Fatalf("code language = %#v", lines[3])
	}
	if lines[4].text != "  fmt.Println(\"ok\")" || lines[4].bg != codeBg {
		t.Fatalf("code line = %#v", lines[4])
	}
}

func TestRespondAskUsesSelectionsAndDefaults(t *testing.T) {
	responses := make(chan askuser.Response, 1)
	m := &Model{
		askReq: &askuser.Request{Questions: []askuser.Question{
			{Header: "Color", Options: []askuser.Option{{Label: "Blue"}, {Label: "Green"}}},
			{Header: "Mode", Options: []askuser.Option{{Label: "Safe"}, {Label: "Fast"}}},
		}, ResponseCh: responses},
		askSel:  []int{1, 0},
		askOpen: state.New(true),
	}
	m.respondAsk(false)
	response := <-responses
	if response.Answers["Color"] != "Green" || response.Answers["Mode"] != "Safe" {
		t.Fatalf("selected response = %#v", response.Answers)
	}
	if m.askReq != nil || m.askOpen.Value() {
		t.Fatal("response should close the modal and clear the request")
	}
}
