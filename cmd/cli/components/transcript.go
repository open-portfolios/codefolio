package components

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
	"github.com/mattn/go-runewidth"
	"github.com/open-portfolios/codefolio/cmd/cli/controller"
)

var (
	accent  = style.HexColor("#8B5CF6")
	muted   = style.HexColor("#9CA3AF")
	primary = style.HexColor("#D1D5DB")
	errorFg = style.HexColor("#EF4444")
	userBar = style.HexColor("#818CF8")
	codeFg  = style.HexColor("#E5E7EB")
	codeBg  = style.HexColor("#1F2937")
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
	return &Transcript{props: props, width: 80}
}

func (t *Transcript) Render(ctx renderer.Context) *renderer.Element {
	content := &renderer.Element{}
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
				current.regions = append(current.regions, clickRegion{rect: renderer.Rect{X: draw.Origin.X + box.X, Y: draw.Origin.Y + box.Y + row, Width: box.Width, Height: 1}, kind: line.kind, messageID: line.messageID, toolID: line.toolID})
			}
		}
		content.SetRect(renderer.Rect{X: draw.Origin.X + box.X, Y: draw.Origin.Y + box.Y, Width: box.Width, Height: len(lines)})
	})
	content.SetHandleMouseFn(func(ev input.KeyEvent) bool {
		current := renderer.Props[*Transcript](content)
		if ev.Mouse.Action != input.MousePress || ev.Mouse.Button != input.MouseLeft {
			return false
		}
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
		return false
	})
	viewport := builtin.CreateViewport(ctx, builtin.ViewportProps{Key: "chat", State: t.props.Viewport, ScrollY: true}, content)
	viewport.SetFlex(1)
	return viewport
}

func (t *Transcript) lines(width int) []visualLine {
	var lines []visualLine
	for _, message := range t.props.Messages {
		switch message.Role {
		case "user":
			lines = append(lines, wrap("| "+message.Content, width, userBar, 0, "", message.ID, "")...)
		case "assistant":
			if message.Thinking != "" {
				prefix := "+ Thinking"
				if message.Streaming && message.Content == "" {
					prefix = spinner[t.props.Frame%len(spinner)] + " Thinking"
				} else if message.ThinkingExpanded {
					prefix = "- Thinking"
				}
				lines = append(lines, visualLine{text: prefix, fg: muted, attrs: style.Italic, kind: "thinking", messageID: message.ID})
				if message.ThinkingExpanded {
					lines = append(lines, wrap("  "+message.Thinking, width, muted, style.Italic, "", message.ID, "")...)
				}
			}
			if message.Content != "" {
				lines = append(lines, markdown(message.Content, width, message.ID)...)
			}
			if message.Error != "" {
				lines = append(lines, wrap("Error: "+message.Error, width, errorFg, 0, "", message.ID, "")...)
			}
			for _, tool := range message.Tools {
				label := toolLabel(tool)
				if !tool.Done {
					label = spinner[t.props.Frame%len(spinner)] + " " + label
				}
				kind := ""
				if tool.Done && showOutput(tool.Name) {
					kind = "tool"
					if tool.Expanded {
						label = "- " + label
					} else {
						label = "+ " + label
					}
				}
				if tool.Done && tool.Elapsed > 0 {
					label += " (" + tool.Elapsed.Round(time.Millisecond).String() + ")"
				}
				lines = append(lines, wrap("  "+label, width, muted, 0, kind, message.ID, tool.ID)...)
				if tool.Done && tool.Expanded && tool.Output != "" {
					fg := muted
					if tool.IsError {
						fg = errorFg
					}
					lines = append(lines, wrap("    "+tool.Output, width, fg, 0, "", message.ID, "")...)
				}
			}
		}
		lines = append(lines, visualLine{})
	}
	if len(lines) == 0 {
		return []visualLine{{text: "Ready.", fg: muted}}
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
			for _, line := range wrap("  "+source, width, codeFg, 0, "", messageID, "") {
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
		result = append(result, wrap(text, width, fg, attrs, "", messageID, "")...)
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
