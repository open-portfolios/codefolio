package tui

import (
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

type ChatModel struct {
	viewport        viewport.Model
	renderer        *glamour.TermRenderer
	width           int
	height          int
	screenY         int
	spinnerFrame    int
	thinkingRegions []thinkingRegion
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

func (c *ChatModel) SetScreenY(y int) {
	c.screenY = y
}

func (c *ChatModel) SetSpinnerFrame(f int) {
	c.spinnerFrame = f
}

func (c *ChatModel) ScreenY() int              { return c.screenY }
func (c *ChatModel) YOffset() int              { return c.viewport.YOffset() }
func (c *ChatModel) VisibleLineCount() int     { return c.viewport.VisibleLineCount() }
func (c *ChatModel) AtBottom() bool            { return c.viewport.AtBottom() }

func (c *ChatModel) ThinkingLineToMsg(line int) (int, bool) {
	for _, r := range c.thinkingRegions {
		if r.collapserLine == line {
			return r.msgIndex, true
		}
	}
	return 0, false
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
	var sb strings.Builder
	c.thinkingRegions = c.thinkingRegions[:0]
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
					for _, tline := range strings.Split(msg.Thinking, "\n") {
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
				errorText := errorStyle.Render("Error: " + msg.Error)
				sb.WriteString(errorText)
				sb.WriteString("\n")
				line++
			}
		default:
			content := lipgloss.NewStyle().Foreground(mutedFg).Render(msg.Content)
			sb.WriteString(content)
			sb.WriteString("\n")
			line += strings.Count(content, "\n") + 1
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

func (c *ChatModel) thinkingHeader(expanded, spinning bool) string {
	content := "+ Thinking"
	if spinning {
		content = spinnerFrames[c.spinnerFrame%len(spinnerFrames)] + " Thinking"
	} else if expanded {
		content = "- Thinking"
	}
	return lipgloss.NewStyle().PaddingLeft(2).Render(thinkingHeaderStyle.Render(content))
}
