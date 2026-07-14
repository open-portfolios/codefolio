package tui

import (
	"encoding/json"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"

	"github.com/open-portfolios/codefolio/internal/domain"
)

type thinkingRegion struct {
	msgIndex      int
	collapserLine int
}

type toolCallRegion struct {
	tcID string
	line int
}

type cachedResult struct {
	content string
	isError bool
}

type ChatModel struct {
	viewport         viewport.Model
	renderer         *glamour.TermRenderer
	width            int
	height           int
	screenY          int
	spinnerFrame     int
	thinkingRegions  []thinkingRegion
	toolCallRegions  []toolCallRegion
	toolCallExpanded map[string]bool
}

func NewChatModel(w, h int) ChatModel {
	vp := viewport.New(
		viewport.WithWidth(w),
		viewport.WithHeight(h),
	)
	return ChatModel{
		viewport:         vp,
		width:            w,
		height:           h,
		toolCallExpanded: make(map[string]bool),
	}
}

func (c *ChatModel) SetSize(w, h int) {
	c.width = w
	c.height = h
	c.viewport.SetWidth(w)
	c.viewport.SetHeight(h)
	c.renderer = nil
}

func (c *ChatModel) SetScreenY(y int) {
	c.screenY = y
}

func (c *ChatModel) SetSpinnerFrame(f int) {
	c.spinnerFrame = f
}

func (c *ChatModel) ScreenY() int          { return c.screenY }
func (c *ChatModel) YOffset() int          { return c.viewport.YOffset() }
func (c *ChatModel) VisibleLineCount() int { return c.viewport.VisibleLineCount() }
func (c *ChatModel) AtBottom() bool        { return c.viewport.AtBottom() }

func (c *ChatModel) ThinkingLineToMsg(line int) (int, bool) {
	for _, r := range c.thinkingRegions {
		if r.collapserLine == line {
			return r.msgIndex, true
		}
	}
	return 0, false
}

func (c *ChatModel) ToolCallLineToID(line int) (string, bool) {
	for _, r := range c.toolCallRegions {
		if r.line == line {
			return r.tcID, true
		}
	}
	return "", false
}

func (c *ChatModel) Rebuild(allMsgs []domain.ChatMessage) {
	c.buildContent(allMsgs)
}

func (c *ChatModel) RebuildAndScroll(allMsgs []domain.ChatMessage) {
	c.buildContent(allMsgs)
	c.viewport.GotoBottom()
}

func (c *ChatModel) View() string {
	return c.viewport.View()
}

func (c *ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)
	return *c, cmd
}

func (c *ChatModel) buildContent(messages []domain.ChatMessage) {
	completedTools := make(map[string]bool)
	toolResultContent := make(map[string]cachedResult)
	for _, msg := range messages {
		if msg.ToolCallID != "" {
			completedTools[msg.ToolCallID] = true
			toolResultContent[msg.ToolCallID] = cachedResult{
				content: msg.Content,
				isError: msg.IsError,
			}
		}
	}

	var sb strings.Builder
	c.thinkingRegions = c.thinkingRegions[:0]
	c.toolCallRegions = c.toolCallRegions[:0]
	w := max(20, c.width-2)
	line := 0

	for msgIdx, msg := range messages {
		switch msg.Role {
		case "system":
			continue
		case "user":
			bar := userMsgBar.Render("")
			body := lipgloss.NewStyle().PaddingLeft(1).MaxWidth(w).Render(msg.Content)
			row := lipgloss.JoinHorizontal(lipgloss.Top, bar, lipgloss.NewStyle().Width(w-1).Render(body))
			sb.WriteString(row)
			sb.WriteString("\n")
			line += strings.Count(row, "\n") + 1
		case "assistant":
			if msg.Thinking != "" {
				sb.WriteString("\n")
				line++

				c.thinkingRegions = append(c.thinkingRegions, thinkingRegion{
					msgIndex:      msgIdx,
					collapserLine: line,
				})

				spinning := msg.Streaming && msg.Content == ""
				sb.WriteString(c.thinkingHeader(msg.ThinkingExpanded, spinning))
				sb.WriteString("\n")
				line++

				if msg.ThinkingExpanded {
					wrapW := max(20, c.width-4)
					for tline := range strings.SplitSeq(msg.Thinking, "\n") {
						lineWrapped := lipgloss.NewStyle().Width(wrapW).MaxWidth(wrapW).PaddingLeft(2).Render(tline)
						wrapped := thinkingStyle.Render(lineWrapped)
						sb.WriteString(wrapped)
						sb.WriteString("\n")
						line += strings.Count(wrapped, "\n") + 1
					}
				}
			}
			body := msg.Content
			if body != "" {
				body = c.renderMarkdown(body)
			}
			sb.WriteString(body)
			sb.WriteString("\n")
			line += strings.Count(body, "\n") + 1
			if msg.Error != "" {
				e := errorStyle.Width(w).Render("Error: " + msg.Error)
				sb.WriteString(e)
				sb.WriteString("\n")
				line += strings.Count(e, "\n") + 1
			}
			for _, tc := range msg.ToolCalls {
				c.renderToolCallLine(&sb, &line, w, tc, completedTools[tc.ID], toolResultContent)
			}
		default:
			continue
		}
	}

	c.viewport.SetContent(sb.String())
}

func (c *ChatModel) renderToolCallLine(sb *strings.Builder, line *int, w int, tc domain.ToolCall, completed bool, results map[string]cachedResult) {
	label := toolSummary(tc)
	if label == "" {
		return
	}

	skip := skipToolResult(tc.Name)

	if !completed {
		content := spinnerFrames[c.spinnerFrame%len(spinnerFrames)] + " " + label
		rendered := lipgloss.NewStyle().
			Foreground(mutedFg).
			PaddingLeft(2).
			Render(content)
		sb.WriteString(rendered)
		sb.WriteString("\n")
		*line += strings.Count(rendered, "\n") + 1
		return
	}

	if skip {
		rendered := lipgloss.NewStyle().
			Foreground(mutedFg).
			PaddingLeft(2).
			Render(label)
		sb.WriteString(rendered)
		sb.WriteString("\n")
		*line += strings.Count(rendered, "\n") + 1
		return
	}

	expanded := c.toolCallExpanded[tc.ID]
	c.toolCallRegions = append(c.toolCallRegions, toolCallRegion{tcID: tc.ID, line: *line})

	prefix := "+ "
	if expanded {
		prefix = "- "
	}
	rendered := lipgloss.NewStyle().
		Foreground(accent).
		PaddingLeft(2).
		Render(prefix + label)
	sb.WriteString(rendered)
	sb.WriteString("\n")
	*line += strings.Count(rendered, "\n") + 1

	if expanded {
		result, ok := results[tc.ID]
		if ok {
			style := lipgloss.NewStyle().PaddingLeft(4).MaxWidth(w)
			if result.isError {
				style = style.Foreground(lipgloss.Color("#EF4444"))
			} else {
				style = style.Foreground(mutedFg)
			}
			for resultLine := range strings.SplitSeq(result.content, "\n") {
				renderedLine := style.Render(resultLine)
				sb.WriteString(renderedLine)
				sb.WriteString("\n")
				*line += strings.Count(renderedLine, "\n") + 1
			}
		}
	}
}

func (c *ChatModel) renderMarkdown(content string) string {
	if c.renderer == nil {
		r, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(max(20, c.width-4)),
		)
		if err != nil {
			return content
		}
		c.renderer = r
	}
	rendered, err := c.renderer.Render(content)
	if err != nil {
		return content
	}
	return rendered
}

func (c *ChatModel) thinkingHeader(expanded, spinning bool) string {
	content := "+ Thinking"
	if spinning {
		content = spinnerFrames[c.spinnerFrame%len(spinnerFrames)] + " Thinking"
	} else if expanded {
		content = "- Thinking"
	}
	return lipgloss.NewStyle().PaddingLeft(2).Render(thinkingHeaderStyle.Render(content))
}

func toolSummary(tc domain.ToolCall) string {
	switch tc.Name {
	case "ReadFile":
		return "Read " + extractField(tc.Input, "file_path")
	case "WriteFile":
		return "Wrote " + extractField(tc.Input, "file_path")
	case "EditFile":
		return "Edited " + extractField(tc.Input, "file_path")
	case "Glob":
		return "Glob " + extractField(tc.Input, "pattern")
	case "Grep":
		return "Grep " + extractField(tc.Input, "pattern")
	case "Bash":
		return "Bash " + extractField(tc.Input, "command")
	}
	return ""
}

func skipToolResult(name string) bool {
	switch name {
	case "ReadFile", "WriteFile", "EditFile", "Grep", "Glob":
		return true
	}
	return false
}

func extractField(inputJSON, key string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(inputJSON), &m); err != nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
