package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

type streamingState int

const (
	streamIdle streamingState = iota
	streamRunning
)

const (
	fixedOverhead = 3
	helpBarH      = 1
	inputMaxH     = 8
	minChatH      = 4
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Model struct {
	width  int
	height int

	chat  ChatModel
	input InputModel

	session *domain.Session

	driver  llm.Driver
	cfg     *conf.Global
	Program *tea.Program

	streaming    streamingState
	spinnerFrame int
	quitting     bool
}

func NewModel(cfg *conf.Global, driver llm.Driver, session *domain.Session) *Model {
	w := 80
	h := 24

	return &Model{
		cfg:     cfg,
		driver:  driver,
		session: session,
		chat:    NewChatModel(w-1, h-4),
		input:   NewInputModel(w),
		width:   w,
		height:  h,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.input.Init(), func() tea.Msg {
		time.Sleep(80 * time.Millisecond)
		return spinnerTickMsg{}
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width == 0 || msg.Height == 0 {
			return m, nil
		}
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		m.chat.Rebuild(m.session.Messages)

	case tea.KeyPressMsg:
		if m.streaming == streamRunning {
			if msg.String() == "esc" {
				return m, nil
			}
			return m, nil
		}

	case inputSentMsg:
		content := strings.TrimSpace(msg.content)
		if content == "" {
			return m, nil
		}
		m.session.AddUserMessage(content)
		m.session.StartAssistantMessage()
		m.streaming = streamRunning
		m.chat.RebuildAndScroll(m.session.Messages)
		return m, StreamLLM(m.Program, m.driver, m.session.ToLLMMessages(), m.cfg.Model)

	case streamDeltaMsg:
		wasAtBottom := m.chat.AtBottom()
		m.session.AppendDelta(string(msg))
		if wasAtBottom {
			m.chat.RebuildAndScroll(m.session.Messages)
		} else {
			m.chat.Rebuild(m.session.Messages)
		}
		return m, nil

	case streamThinkingMsg:
		wasAtBottom := m.chat.AtBottom()
		m.session.AppendThinkingDelta(string(msg))
		if wasAtBottom {
			m.chat.RebuildAndScroll(m.session.Messages)
		} else {
			m.chat.Rebuild(m.session.Messages)
		}
		return m, nil

	case streamDoneMsg:
		m.session.FinishAssistantMessage()
		m.streaming = streamIdle
		m.chat.RebuildAndScroll(m.session.Messages)
		return m, nil

	case streamErrMsg:
		m.session.FinishAssistantMessage()
		if last := m.session.LastMessage(); last != nil {
			last.Content = fmt.Sprintf("Error: %s", msg.Err.Error())
		}
		m.streaming = streamIdle
		m.chat.Rebuild(m.session.Messages)
		return m, nil

	case spinnerTickMsg:
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		if m.streaming == streamRunning {
			m.chat.SetSpinnerFrame(m.spinnerFrame)
			m.chat.Rebuild(m.session.Messages)
		}
		return m, func() tea.Msg {
			time.Sleep(80 * time.Millisecond)
			return spinnerTickMsg{}
		}

	case tea.MouseWheelMsg:
		var cmd tea.Cmd
		m.chat, cmd = m.chat.Update(msg)
		return m, cmd

	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			m.handleChatClick(msg.X, msg.Y)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) View() tea.View {
	if m.quitting {
		return tea.NewView("Goodbye!\n")
	}

	workDir, _ := os.Getwd()
	header := renderHeader(m.cfg.Model, workDir, m.width)
	chatView := m.chat.View()
	inputView := m.input.View()

	var sidebar string
	hasSidebar := m.width >= 120
	if hasSidebar {
		sidebar = renderSidebar(m.session, 28)
	}

	var body string
	if hasSidebar {
		body = lipgloss.JoinHorizontal(lipgloss.Top,
			chatView,
			sidebar,
		)
	} else {
		body = chatView
	}

	m.chat.SetScreenY(2)

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		body,
		"",
		inputView,
	))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m *Model) handleChatClick(mx, my int) {
	hasSidebar := m.width >= 120
	if hasSidebar {
		chatW := m.width - 30
		if mx >= chatW {
			return
		}
	}

	localY := my - m.chat.ScreenY()
	if localY < 0 || localY >= m.chat.VisibleLineCount() {
		return
	}

	contentLine := localY + m.chat.YOffset()
	msgIdx, ok := m.chat.ThinkingLineToMsg(contentLine)
	if !ok {
		return
	}

	if msgIdx < len(m.session.Messages) {
		m.session.Messages[msgIdx].ThinkingExpanded = !m.session.Messages[msgIdx].ThinkingExpanded
		m.chat.Rebuild(m.session.Messages)
	}
}

func (m *Model) updateLayout() {
	sw := 0
	if m.width >= 120 {
		sw = 30
	}
	mainW := m.width - sw

	inputH := min(max(m.height-fixedOverhead-helpBarH-minChatH, 1), inputMaxH)
	chatH := max(m.height-fixedOverhead-inputH-helpBarH, minChatH)

	m.chat.SetSize(mainW, chatH)
	m.input.SetSize(m.width, inputH)
}
