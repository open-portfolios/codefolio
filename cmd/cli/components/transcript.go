package components

import (
	"encoding/json"
	"strings"

	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
	"github.com/mattn/go-runewidth"
	"github.com/open-portfolios/codefolio/cmd/cli/controller"
)

var (
	accent  = Theme.Accent
	muted   = Theme.TextMuted
	primary = Theme.Text
	errorFg = Theme.Error
	codeFg  = Theme.Text
	codeBg  = Theme.BackgroundElement
	spinner = []string{"|", "/", "-", "\\"}
)

type TranscriptProps struct {
	Messages         []*controller.Message
	Viewport         *state.State[builtin.ViewportState]
	Frame            int
	OnToggleThinking func(string)
	OnToggleTool     func(string, string)
}

type Transcript struct {
	props   TranscriptProps
	width   int
	regions []clickRegion
}
type clickRegion struct {
	rect      renderer.Rect
	kind      string
	messageID string
	toolID    string
}
type visualLine struct {
	text      string
	fg        style.Color
	bg        style.Color
	attrs     style.Attr
	kind      string
	messageID string
	toolID    string
}

func NewTranscript(ctx renderer.Context, props TranscriptProps, children ...renderer.Component) *Transcript {
	return &Transcript{props: props, width: transcriptWidth(ctx)}
}

func (t *Transcript) Render(ctx renderer.Context) *renderer.Element {
	content := &renderer.Element{}
	var viewport *renderer.Element
	content.SetKey("transcript")
	content.SetTag("transcript")
	content.SetProps(t)
	content.SetLayoutHeight(len(t.lines(max(t.width, 20))))
	content.SetPaintFn(func(draw renderer.DrawContext, box renderer.Rect) {
		current := renderer.Props[*Transcript](content)
		width := max(box.Width, 20)
		if width != current.width {
			current.width = width
			ctx.MarkDirty()
		}
		lines := current.lines(width)
		current.regions = current.regions[:0]
		for row, line := range lines {
			if row >= box.Height {
				break
			}
			if line.bg != (style.Color{}) {
				draw.Fill(renderer.Rect{X: box.X, Y: box.Y + row, Width: box.Width, Height: 1}, renderer.Cell{Rune: ' ', Style: style.Style{Bg: line.bg}})
			}
			draw.WriteStringWide(box.X, box.Y+row, line.text, style.Style{Fg: line.fg, Bg: line.bg, Attrs: line.attrs})
			if line.kind != "" {
				current.regions = append(current.regions, clickRegion{rect: renderer.Rect{X: draw.Origin.X + box.X, Y: draw.Origin.Y + box.Y + row, Width: runewidth.StringWidth(line.text), Height: 1}, kind: line.kind, messageID: line.messageID, toolID: line.toolID})
			}
		}
		content.SetRect(renderer.Rect{X: draw.Origin.X + box.X, Y: draw.Origin.Y + box.Y, Width: box.Width, Height: len(lines)})
	})
	content.SetHandleMouseFn(func(ev input.KeyEvent) bool {
		current := renderer.Props[*Transcript](content)
		if ev.Mouse.Action == input.MouseWheel {
			if current.props.Viewport == nil || ev.Mouse.Scroll == 0 {
				return false
			}
			viewportState := current.props.Viewport.Value()
			if viewportState.ContentHeight <= viewportState.ViewportHeight {
				return false
			}
			viewportState.OffsetY -= ev.Mouse.Scroll
			viewportState.Clamp()
			viewportState.FollowEnd = viewportState.AtEnd()
			current.props.Viewport.Set(viewportState)
			ctx.Focus(viewport)
			return true
		}
		if ev.Mouse.Action != input.MousePress || ev.Mouse.Button != input.MouseLeft {
			return false
		}
		ctx.Focus(viewport)
		for _, region := range current.regions {
			if contains(region.rect, ev.Mouse.X, ev.Mouse.Y) {
				if region.kind == "thinking" {
					current.props.OnToggleThinking(region.messageID)
				} else {
					current.props.OnToggleTool(region.messageID, region.toolID)
				}
				return true
			}
		}
		return true
	})
	viewport = builtin.CreateViewport(ctx, builtin.ViewportProps{Key: "chat", State: t.props.Viewport, ScrollY: true, StickToBottom: true}, content)
	viewport.SetFlex(1)
	return viewport
}

// transcriptWidth mirrors Shell's sidebar split and horizontal inset so the
// viewport height is measured using the same width used for painting.
func transcriptWidth(ctx renderer.Context) int {
	width, _ := ctx.Size()
	if width > 120 {
		width -= 42
	}
	return max(width-4, 20)
}

func (t *Transcript) lines(width int) []visualLine {
	var lines []visualLine
	items := projectTimeline(t.props.Messages)
	for index, item := range items {
		switch item.Kind {
		case TimelineUserMessage:
			for _, line := range userMessageLines(item.Content, width, item.MessageID) {
				line.bg = Theme.BackgroundPanel
				lines = append(lines, line)
			}
		case TimelineAssistantMarkdown:
			lines = append(lines, markdown(item.Content, width, item.MessageID)...)
		case TimelineThinking:
			prefix := "+ Thought"
			if item.Streaming {
				prefix = spinner[t.props.Frame%len(spinner)] + " Thinking"
			} else if item.Expanded {
				prefix = "- Thought"
			}
			lines = append(lines, visualLine{text: "   " + prefix, fg: Theme.Warning, attrs: style.Italic, kind: "thinking", messageID: item.MessageID})
			if item.Expanded {
				lines = append(lines, wrap("     "+item.Content, width, muted, style.Italic, "", item.MessageID, "")...)
			}
		case TimelineToolActivity:
			fg, attrs := toolStyle(item.Tool)
			kind := ""
			if item.Tool.Done && showOutput(item.Tool.Name) {
				kind = "tool"
			}
			lines = append(lines, wrap("   "+toolSummary(item.Tool, t.props.Frame), width, fg, attrs, kind, item.MessageID, item.ToolID)...)
		case TimelineToolOutput:
			fg := Theme.TextMuted
			if item.Tool.IsError && item.Tool.Outcome != "permission_denied" && item.Tool.Outcome != "permission_aborted" {
				fg = Theme.Error
			} else if item.Tool.Outcome == "permission_denied" || item.Tool.Outcome == "permission_aborted" {
				fg = Theme.Warning
			}
			for _, line := range previewOutput(item.Content, width, 10, item.MessageID, item.ToolID, fg) {
				line.bg = Theme.BackgroundPanel
				lines = append(lines, line)
			}
		case TimelineError:
			lines = append(lines, wrap("   Error: "+item.Content, width, errorFg, 0, "", item.MessageID, "")...)
		case TimelineTurnMeta:
			lines = append(lines, visualLine{text: "   ▣  " + strings.ToUpper(item.Content[:1]) + item.Content[1:], fg: Theme.Error, messageID: item.MessageID})
		}
		if index < len(items)-1 && item.MessageID != items[index+1].MessageID {
			lines = append(lines, visualLine{})
		}
	}
	if len(lines) == 0 {
		banner := CodefolioBanner()
		lines = make([]visualLine, 0, len(banner)+2)
		for _, row := range banner {
			lines = append(lines, visualLine{text: row, fg: Theme.Primary, attrs: style.Bold})
		}
		lines = append(lines, visualLine{})
		lines = append(lines, visualLine{text: "Ready when you are.", fg: muted})
		return lines
	}
	return trimTrailingBlankLines(lines)
}

func userMessageLines(content string, width int, messageID string) []visualLine {
	wrapped := wrap(content, max(width-2, 1), Theme.Text, 0, "", messageID, "")
	lines := make([]visualLine, 0, len(wrapped)+2)
	lines = append(lines, visualLine{messageID: messageID})
	lines = append(lines, wrapped...)
	lines = append(lines, visualLine{messageID: messageID})
	for i := range lines {
		lines[i].text = "┃ " + lines[i].text
	}
	return lines
}

func trimTrailingBlankLines(lines []visualLine) []visualLine {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1].text) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func wrap(text string, width int, fg style.Color, attrs style.Attr, kind, messageID, toolID string) []visualLine {
	var result []visualLine
	for source := range strings.SplitSeq(text, "\n") {
		if source == "" {
			result = append(result, visualLine{fg: fg, attrs: attrs, kind: kind, messageID: messageID, toolID: toolID})
			continue
		}
		var line strings.Builder
		lineWidth := 0
		for _, r := range source {
			rw := max(runewidth.RuneWidth(r), 1)
			if lineWidth+rw > width && line.Len() > 0 {
				result = append(result, visualLine{text: line.String(), fg: fg, attrs: attrs, kind: kind, messageID: messageID, toolID: toolID})
				line.Reset()
				lineWidth = 0
			}
			line.WriteRune(r)
			lineWidth += rw
		}
		result = append(result, visualLine{text: line.String(), fg: fg, attrs: attrs, kind: kind, messageID: messageID, toolID: toolID})
	}
	return result
}
func markdown(content string, width int, messageID string) []visualLine {
	var result []visualLine
	inCode := false
	for source := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(source)
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			if inCode && len(trimmed) > 3 {
				result = append(result, visualLine{text: "  " + strings.TrimSpace(trimmed[3:]), fg: muted, bg: codeBg, attrs: style.Dim, messageID: messageID})
			}
			continue
		}
		if inCode {
			for _, line := range wrap("   "+source, width, codeFg, 0, "", messageID, "") {
				line.bg = codeBg
				result = append(result, line)
			}
			continue
		}
		fg, attrs := primary, style.Attr(0)
		text := source
		switch {
		case strings.HasPrefix(trimmed, "### "), strings.HasPrefix(trimmed, "## "), strings.HasPrefix(trimmed, "# "):
			text = strings.TrimLeft(trimmed, "# ")
			fg, attrs = accent, style.Bold
		case strings.HasPrefix(trimmed, "> "):
			text = "| " + strings.TrimPrefix(trimmed, "> ")
			fg, attrs = muted, style.Italic
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			text = "• " + trimmed[2:]
		default:
			text = inlineCode(text)
		}
		result = append(result, wrap("   "+text, width, fg, attrs, "", messageID, "")...)
	}
	return result
}
func inlineCode(text string) string {
	if strings.Count(text, "`") < 2 {
		return text
	}
	return strings.ReplaceAll(text, "`", "'")
}
func toolLabel(tool *controller.Tool) string {
	verb := map[string]string{"ReadFile": "Read", "WriteFile": "Wrote", "EditFile": "Edited", "Glob": "Glob", "Grep": "Grep", "Bash": "Bash", "AskUserQuestion": "Question"}[tool.Name]
	if verb == "" {
		verb = tool.Name
	}
	field := "file_path"
	if tool.Name == "Glob" || tool.Name == "Grep" {
		field = "pattern"
	}
	if tool.Name == "Bash" {
		field = "command"
	}
	value := extract(tool.Input, field)
	if value == "" {
		return verb
	}
	return verb + " " + value
}
func showOutput(name string) bool { return name == "Bash" || name == "AskUserQuestion" }

func previewOutput(output string, width, maxLines int, messageID, toolID string, fg style.Color) []visualLine {
	lines := wrap("     "+output, width, fg, 0, "", messageID, toolID)
	if len(lines) <= maxLines {
		return lines
	}
	preview := append([]visualLine(nil), lines[:maxLines]...)
	preview = append(preview, visualLine{text: "     … output truncated", fg: Theme.TextMuted, attrs: style.Dim, messageID: messageID, toolID: toolID})
	return preview
}
func extract(raw, key string) string {
	var values map[string]any
	if json.Unmarshal([]byte(raw), &values) != nil {
		return ""
	}
	value, _ := values[key].(string)
	return value
}
func contains(rect renderer.Rect, x, y int) bool {
	return x >= rect.X && y >= rect.Y && x < rect.X+rect.Width && y < rect.Y+rect.Height
}
