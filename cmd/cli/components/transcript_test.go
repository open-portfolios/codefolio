package components

import (
	"testing"

	"github.com/open-portfolios/codefolio/cmd/cli/controller"
)

func TestWrapPreservesContentAndWidth(t *testing.T) {
	lines := wrap("hello world", 5, primary, 0, "", "message", "")
	if len(lines) != 3 {
		t.Fatalf("line count = %d, want 3", len(lines))
	}
	if lines[0].text != "hello" || lines[1].text != " worl" || lines[2].text != "d" {
		t.Fatalf("wrapped lines = %#v", lines)
	}
}

func TestToolLabelHandlesKnownAndUnknownTools(t *testing.T) {
	if got := toolLabel(&controller.Tool{Name: "Bash", Input: `{"command":"go test ./..."}`}); got != "Bash go test ./..." {
		t.Fatalf("Bash label = %q", got)
	}
	if got := toolLabel(&controller.Tool{Name: "CustomTool"}); got != "CustomTool" {
		t.Fatalf("unknown label = %q", got)
	}
}
