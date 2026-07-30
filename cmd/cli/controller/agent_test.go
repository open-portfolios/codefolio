package controller

import (
	"testing"

	"github.com/open-portfolios/codefolio/internal/domain"
)

func TestApplyToolCallUpdatesExistingToolInput(t *testing.T) {
	c := &Controller{
		messages:  []*Message{{Role: "assistant", Tools: []*Tool{{ID: "tool-1", Name: "Bash"}}}},
		streaming: StreamRunning,
		runID:     1,
		invalidate: func() {
		},
	}
	v := visitor{controller: c, runID: 1}

	if err := v.applyToolCall(domain.ToolCallEvent{ID: "tool-1", Name: "Bash", Input: `{"command":"go test ./..."}`}); err != nil {
		t.Fatalf("applyToolCall returned error: %v", err)
	}
	if len(c.messages[0].Tools) != 1 {
		t.Fatalf("tool count = %d, want 1", len(c.messages[0].Tools))
	}
	if got := c.messages[0].Tools[0].Input; got != `{"command":"go test ./..."}` {
		t.Fatalf("tool input = %q", got)
	}
}

func TestStartAssistantSegmentSeparatesToolContinuation(t *testing.T) {
	c := &Controller{
		messages: []*Message{{ID: "assistant-1", Role: "assistant", Tools: []*Tool{{ID: "tool-1", Name: "Bash", Done: true}}}},
		invalidate: func() {
		},
	}

	c.StartAssistantSegment()

	if len(c.messages) != 2 {
		t.Fatalf("message count = %d, want 2", len(c.messages))
	}
	if got := c.messages[1]; got.ID != "assistant-2" || got.Role != "assistant" || !got.Streaming {
		t.Fatalf("continuation segment = %#v, want streaming assistant-2", got)
	}
}
