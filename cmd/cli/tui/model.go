package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/pkg/llm"
)

type streamingState int

const (
	streamIdle streamingState = iota
	streamRunning
)

type Model struct {
	width  int
	height int

	chat  ChatModel
	input InputModel

	session *Session

	driver  llm.Driver
	cfg     *conf.Global
	Program *tea.Program

	streaming streamingState
	quitting  bool
}

func NewModel(cfg *conf.Global, driver llm.Driver) *Model {
	w := 80
	h := 24

	session := NewSession()
	session.AddSystemMessage("you're a helpful assistant")

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
	return m.input.Init()
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
		m.chat.Rebuild(m.session.Messages)
		return m, StreamLLM(m.Program, m.driver, m.session.ToLLMMessages(), m.cfg.Model)

	case streamDeltaMsg:
		m.session.AppendDelta(string(msg))
		m.chat.RebuildAndScroll(m.session.Messages)
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

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		body,
		inputView,
	))
	v.AltScreen = true
	return v
}

func (m *Model) updateLayout() {
	sw := 0
	if m.width >= 120 {
		sw = 30
	}
	mainW := m.width - sw

	chatH := m.height - 4
	if chatH < 4 {
		chatH = 4
	}

	m.chat.SetSize(mainW, chatH)
	m.input.SetSize(m.width)
}

func (s *Session) AddSystemMessage(content string) {
	s.msgSeq++
	msg := ChatMessage{
		ID:        sysMsgID(s.msgSeq),
		Role:      "system",
		Content:   content,
		Timestamp: time.Now(),
	}
	s.Messages = append(s.Messages, msg)
}

func sysMsgID(seq int) string { return "s-" + itoa(seq) }
