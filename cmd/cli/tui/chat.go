package tui

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

type ChatModel struct {
	viewport viewport.Model
	renderer *glamour.TermRenderer
	width    int
	height   int
}

func NewChatModel(w, h int) ChatModel {
	vp := viewport.New(
		viewport.WithWidth(w),
		viewport.WithHeight(h),
	)
	return ChatModel{
		viewport: vp,
		width:    w,
		height:   h,
	}
}

func (c *ChatModel) SetSize(w, h int) {
	c.width = w
	c.height = h
	c.viewport.SetWidth(w)
	c.viewport.SetHeight(h)
	c.renderer = nil
}

func (c *ChatModel) Rebuild(allMsgs []ChatMessage) {
	c.buildContent(allMsgs)
}

func (c *ChatModel) RebuildAndScroll(allMsgs []ChatMessage) {
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

func (c *ChatModel) buildContent(messages []ChatMessage) {
	var sb strings.Builder
	w := max(20, c.width-2)

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			continue
		case "user":
			bar := userMsgBar.Render("")
			body := lipgloss.NewStyle().PaddingLeft(1).MaxWidth(w).Render(msg.Content)
			row := lipgloss.JoinHorizontal(lipgloss.Top, bar, lipgloss.NewStyle().Width(w-1).Render(body))
			sb.WriteString(row)
			sb.WriteString("\n")
		case "assistant":
			body := msg.Content
			if body != "" {
				body = c.renderMarkdown(body)
			}
			sb.WriteString(body)
			sb.WriteString("\n")
		default:
			sb.WriteString(lipgloss.NewStyle().Foreground(mutedFg).Render(msg.Content))
			sb.WriteString("\n")
		}
	}

	c.viewport.SetContent(sb.String())
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
