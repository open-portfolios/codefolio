package components

import (
	"strings"
	"testing"

	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
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

func TestProjectTimelinePlacesToolContinuationAfterToolActivity(t *testing.T) {
	messages := []*controller.Message{{
		ID:       "assistant-1",
		Role:     "assistant",
		Content:  "I will inspect it.",
		Thinking: "Inspecting files",
		Tools: []*controller.Tool{{
			ID: "tool-1", Name: "Bash", Done: true, Expanded: true, Output: "ok",
		}},
	}, {
		ID:      "assistant-2",
		Role:    "assistant",
		Content: "The command completed.",
	}}

	items := projectTimeline(messages)
	if len(items) != 5 {
		t.Fatalf("item count = %d, want 5", len(items))
	}
	if items[0].Kind != TimelineThinking || items[1].Kind != TimelineAssistantMarkdown || items[2].Kind != TimelineToolActivity || items[3].Kind != TimelineToolOutput || items[4].Kind != TimelineAssistantMarkdown {
		t.Fatalf("unexpected timeline item kinds: %#v", items)
	}
	if items[4].Content != "The command completed." {
		t.Fatalf("tool continuation = %q, want post-tool assistant content", items[4].Content)
	}
}

func TestExpandableToolHeaderIsClickableAcrossItsText(t *testing.T) {
	var toggled []string
	transcript := NewTranscript(renderer.Context{MarkDirtyFn: func() {}}, TranscriptProps{
		Messages: []*controller.Message{{
			ID: "assistant-1", Role: "assistant", Tools: []*controller.Tool{{
				ID: "tool-1", Name: "Bash", Input: `{"command":"go test ./..."}`, Done: true, Output: "ok",
			}},
		}},
		Viewport: state.New(builtin.ViewportState{FollowEnd: true}),
		OnToggleTool: func(messageID, toolID string) {
			toggled = append(toggled, messageID+":"+toolID)
		},
	})

	root := transcript.Render(renderer.Context{MarkDirtyFn: func() {}})
	root.Draw(renderer.NewCellBuffer(80, 10), 0, 0, 80, 10)
	content := root.Children()[0]
	if len(transcript.regions) != 1 {
		t.Fatalf("click regions = %#v, want one tool header", transcript.regions)
	}
	region := transcript.regions[0]
	for _, x := range []int{region.rect.X, region.rect.X + region.rect.Width/2, region.rect.X + region.rect.Width - 1} {
		if !content.HandleMouse(input.KeyEvent{Type: input.EventMouse, Mouse: input.MouseEvent{X: x, Y: region.rect.Y, Button: input.MouseLeft, Action: input.MousePress}}) {
			t.Fatalf("click at x=%d was not handled", x)
		}
	}
	if got, want := toggled, []string{"assistant-1:tool-1", "assistant-1:tool-1", "assistant-1:tool-1"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("toggle calls = %#v, want %#v", got, want)
	}
	if content.HandleMouse(input.KeyEvent{Type: input.EventMouse, Mouse: input.MouseEvent{X: region.rect.X + region.rect.Width, Y: region.rect.Y, Button: input.MouseLeft, Action: input.MousePress}}) {
		t.Fatal("click outside the tool header must not toggle it")
	}
}

func TestPreviewOutputTruncatesLongToolOutput(t *testing.T) {
	lines := previewOutput("one\ntwo\nthree", 80, 2, "message", "tool", muted)
	if len(lines) != 3 {
		t.Fatalf("line count = %d, want 3", len(lines))
	}
	if lines[2].text != "     … output truncated" {
		t.Fatalf("truncation label = %q", lines[2].text)
	}
}

func TestTranscriptAddsGapsBetweenMessagesButNotAfterTheLast(t *testing.T) {
	transcript := &Transcript{props: TranscriptProps{Messages: []*controller.Message{
		{ID: "user-1", Role: "user", Content: "hello"},
		{ID: "assistant-1", Role: "assistant", Content: "hi"},
	}}}
	lines := transcript.lines(80)
	if len(lines) < 3 {
		t.Fatalf("line count = %d, want message gap", len(lines))
	}
	assistant := -1
	for i, line := range lines {
		if strings.Contains(line.text, "hi") {
			assistant = i
			break
		}
	}
	if assistant < 1 || lines[assistant-1].text != "" {
		t.Fatalf("transcript must add one blank line between messages: %#v", lines)
	}
	if lines[len(lines)-1].text == "" {
		t.Fatal("transcript must not add an implicit trailing blank line")
	}
}

func TestTranscriptDoesNotGapAssistantPartsWithinTheSameMessage(t *testing.T) {
	transcript := &Transcript{props: TranscriptProps{Messages: []*controller.Message{{
		ID:       "assistant-1",
		Role:     "assistant",
		Thinking: "Inspecting",
		Content:  "Complete.",
	}}}}
	lines := transcript.lines(80)
	if len(lines) != 2 {
		t.Fatalf("line count = %d, want two adjacent assistant parts", len(lines))
	}
	if lines[0].text == "" || lines[1].text == "" {
		t.Fatalf("assistant parts must not be separated by an implicit gap: %#v", lines)
	}
}

func TestInterruptedTurnIsSeparatedFromTheAssistantMessage(t *testing.T) {
	transcript := &Transcript{props: TranscriptProps{Messages: []*controller.Message{
		{ID: "user-1", Role: "user", Content: "Make a plan."},
		{ID: "assistant-2", Role: "assistant", Content: "I will inspect the workspace.", Error: "interrupted"},
	}}}
	lines := transcript.lines(80)

	interrupt := -1
	for i, line := range lines {
		if strings.Contains(line.text, "▣  Interrupted") {
			interrupt = i
			break
		}
	}
	if interrupt < 1 {
		t.Fatalf("interrupt event missing from timeline: %#v", lines)
	}
	if lines[interrupt].fg != Theme.Error || lines[interrupt].attrs != 0 {
		t.Fatalf("interrupt style = %#v, want bright error red without dim", lines[interrupt])
	}
	if strings.TrimSpace(lines[interrupt-1].text) != "" {
		t.Fatalf("interrupt must have one blank line above it: %#v", lines)
	}
	if interrupt != len(lines)-1 {
		t.Fatalf("interrupt must not add a trailing blank line: %#v", lines)
	}
}

func TestTranscriptTrimsTrailingMarkdownNewlines(t *testing.T) {
	transcript := &Transcript{props: TranscriptProps{Messages: []*controller.Message{{
		ID:      "assistant-1",
		Role:    "assistant",
		Content: "Complete.\n\n",
	}}}}
	lines := transcript.lines(80)
	if len(lines) == 0 || strings.TrimSpace(lines[len(lines)-1].text) == "" {
		t.Fatalf("trailing blank lines were not trimmed: %#v", lines)
	}
}

func TestUserMessageRailCoversEveryVisualLine(t *testing.T) {
	lines := userMessageLines("first\nsecond", 20, "user-1")
	if len(lines) != 4 {
		t.Fatalf("line count = %d, want two content lines plus vertical padding", len(lines))
	}
	for i, line := range lines {
		if !strings.HasPrefix(line.text, "┃ ") {
			t.Fatalf("line %d = %q, want user rail", i, line.text)
		}
	}
	if strings.TrimSpace(strings.TrimPrefix(lines[0].text, "┃")) != "" || strings.TrimSpace(strings.TrimPrefix(lines[len(lines)-1].text, "┃")) != "" {
		t.Fatalf("user message must have blank rail padding above and below: %#v", lines)
	}
}

func TestTimelineStaysAnchoredAboveComposerAsMessagesGrow(t *testing.T) {
	for _, count := range []int{1, 4, 8} {
		messages := make([]*controller.Message, 0, count)
		for range count {
			messages = append(messages, &controller.Message{ID: "assistant", Role: "assistant", Content: "message"})
		}
		viewport := state.New(builtin.ViewportState{FollowEnd: true})
		transcript := NewTranscript(renderer.Context{MarkDirtyFn: func() {}}, TranscriptProps{Messages: messages, Viewport: viewport})
		buffer := renderer.NewCellBuffer(40, 10)
		transcript.Render(renderer.Context{MarkDirtyFn: func() {}}).Draw(buffer, 0, 0, 40, 10)

		last := -1
		for y := range 10 {
			if strings.Contains(stringRow(buffer, y), "message") {
				last = y
			}
		}
		if last != 9 {
			t.Fatalf("%d messages end at row %d, want viewport bottom", count, last)
		}
	}
}

func TestTranscriptMeasuresMultilineParagraphsAtShellWidth(t *testing.T) {
	ctx := renderer.Context{MarkDirtyFn: func() {}, SizeFn: func() (int, int) { return 121, 30 }}
	transcript := NewTranscript(ctx, TranscriptProps{
		Messages: []*controller.Message{{
			ID:   "assistant-1",
			Role: "assistant",
			Content: strings.Repeat("First paragraph contains enough words to wrap at the rendered timeline width. ", 3) + "\n\n" +
				strings.Repeat("Second paragraph also wraps over several visual lines in the real session shell. ", 3) + "\n\n" +
				"Final paragraph must remain anchored above the composer.",
		}},
		Viewport: state.New(builtin.ViewportState{FollowEnd: true}),
	})
	root := transcript.Render(ctx)
	content := root.Children()[0]
	wantWidth := 75 // 121 columns - 42-column sidebar - four-column main inset.
	if got, want := content.LayoutHeight(), len(transcript.lines(wantWidth)); got != want {
		t.Fatalf("content height = %d, want %d visual lines at shell width %d", got, want, wantWidth)
	}
}

func stringRow(buffer *renderer.CellBuffer, y int) string {
	line := make([]rune, buffer.Width)
	for x := range buffer.Width {
		line[x] = buffer.Cells[y*buffer.Width+x].Rune
		if line[x] == 0 || line[x] == -1 {
			line[x] = ' '
		}
	}
	return string(line)
}
