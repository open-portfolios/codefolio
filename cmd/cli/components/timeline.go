package components

import (
	"strings"
	"time"

	"github.com/cylixlee/tux/style"
	"github.com/open-portfolios/codefolio/cmd/cli/controller"
)

type TimelineItemKind uint8

const (
	TimelineUserMessage TimelineItemKind = iota
	TimelineAssistantMarkdown
	TimelineThinking
	TimelineToolActivity
	TimelineToolOutput
	TimelineError
	TimelineTurnMeta
	TimelineContextNotice
)

// TimelineItem is the CLI-only presentation model between Agent messages and
// terminal lines. It keeps UI grouping out of the shared Agent domain.
type TimelineItem struct {
	Kind      TimelineItemKind
	MessageID string
	ToolID    string
	Content   string
	Profile   string
	Tool      *controller.Tool
	Streaming bool
	Expanded  bool
}

func projectTimeline(messages []*controller.Message) []TimelineItem {
	items := make([]TimelineItem, 0, len(messages)*2)
	for _, message := range messages {
		switch message.Role {
		case "user":
			items = append(items, TimelineItem{Kind: TimelineUserMessage, MessageID: message.ID, Content: message.Content, Profile: message.Profile})
		case "assistant":
			if message.Thinking != "" {
				items = append(items, TimelineItem{Kind: TimelineThinking, MessageID: message.ID, Content: message.Thinking, Streaming: message.Streaming, Expanded: message.ThinkingExpanded})
			}
			if message.Content != "" {
				items = append(items, TimelineItem{Kind: TimelineAssistantMarkdown, MessageID: message.ID, Content: message.Content})
			}
			for _, tool := range message.Tools {
				items = append(items, TimelineItem{Kind: TimelineToolActivity, MessageID: message.ID, ToolID: tool.ID, Tool: tool})
				if tool.Done && tool.Expanded && tool.Output != "" && showOutput(tool.Name) {
					items = append(items, TimelineItem{Kind: TimelineToolOutput, MessageID: message.ID, ToolID: tool.ID, Tool: tool, Content: tool.Output})
				}
			}
			if message.Error != "" {
				kind := TimelineError
				messageID := message.ID
				if message.Error == "interrupted" {
					kind = TimelineTurnMeta
					// Interrupt is an event after the assistant turn, not part of it.
					// Give it its own presentation identity so Transcript places the
					// normal message gap above it.
					messageID += ":interrupt"
				}
				items = append(items, TimelineItem{Kind: kind, MessageID: messageID, Content: message.Error})
			}
		case "context":
			items = append(items, TimelineItem{Kind: TimelineContextNotice, MessageID: message.ID, Content: message.Content})
		}
	}
	return items
}

func toolGlyph(name string) string {
	if isMCPTool(name) {
		return "◇"
	}
	switch name {
	case "ReadFile":
		return "→"
	case "WriteFile", "EditFile":
		return "✎"
	case "Glob", "Grep":
		return "✱"
	case "Bash":
		return "$"
	case "AskUserQuestion":
		return "?"
	default:
		return "•"
	}
}

func toolStyle(tool *controller.Tool) (style.Color, style.Attr) {
	if tool.Outcome == "permission_denied" || tool.Outcome == "permission_aborted" {
		return Theme.Warning, 0
	}
	if tool.IsError {
		return Theme.Error, 0
	}
	if !tool.Done {
		return Theme.Primary, 0
	}
	return Theme.TextMuted, 0
}

func toolSummary(tool *controller.Tool, frame int) string {
	label := toolGlyph(tool.Name) + " " + toolLabel(tool)
	if !tool.Done {
		return spinner[frame%len(spinner)] + " " + label
	}
	if tool.Outcome == "permission_denied" {
		label = "⊘ Denied " + toolLabel(tool)
	}
	if tool.Outcome == "permission_aborted" {
		label = "⊘ Approval cancelled " + toolLabel(tool)
	}
	if showOutput(tool.Name) {
		if tool.Expanded {
			label = "- " + label
		} else {
			label = "+ " + label
		}
	}
	if tool.Elapsed > 0 {
		label += " · " + tool.Elapsed.Round(time.Millisecond).String()
	}
	return label
}

func isMCPTool(name string) bool { return strings.HasPrefix(name, "mcp__") }
