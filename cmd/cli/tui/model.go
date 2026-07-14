package tui

import (
	"context"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
	"github.com/open-portfolios/codefolio/internal/prompt"
	"github.com/open-portfolios/codefolio/internal/svc"
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

	driver   llm.Driver
	agent    *svc.Agent
	registry *tools.Registry
	cfg      *conf.Global
	Program  *tea.Program

	streaming    streamingState
	spinnerFrame int
	quitting     bool
	cancelStream context.CancelFunc

	pendingToolCalls map[string]int
	askUserCh        <-chan askuser.Request

	askReq  *askuser.Request
	askQ    int
	askCur  int
	askSels []int

	inputTokens  int64
	outputTokens int64
}

type questionRequestMsg struct {
	Request askuser.Request
}

func NewModel(cfg *conf.Global, driver llm.Driver, agent *svc.Agent, session *domain.Session, registry *tools.Registry, askUserCh chan askuser.Request) *Model {
	w := 80
	h := 24

	workDir, _ := os.Getwd()
	env := prompt.DetectEnvironment(workDir)
	env.Model = cfg.Model
	sysPrompt := prompt.BuildSystemPrompt(env, prompt.BuildOptions{Model: cfg.Model})
	session.AddSystemMessage(sysPrompt)

	return &Model{
		cfg:              cfg,
		driver:           driver,
		agent:            agent,
		session:          session,
		registry:         registry,
		chat:             NewChatModel(w-1, h-4),
		input:            NewInputModel(w),
		width:            w,
		height:           h,
		pendingToolCalls: make(map[string]int),
		askUserCh:        askUserCh,
	}
}

func (m *Model) Init() tea.Cmd {
	go m.askUserLoop()
	return tea.Batch(m.input.Init(), func() tea.Msg {
		time.Sleep(80 * time.Millisecond)
		return spinnerTickMsg{}
	})
}

func (m *Model) askUserLoop() {
	for req := range m.askUserCh {
		m.Program.Send(questionRequestMsg{Request: req})
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.askReq != nil {
		return m.updateAskMode(msg)
	}

	switch msg := msg.(type) {
	case questionRequestMsg:
		req := msg.Request
		m.askReq = &req
		m.askQ = 0
		m.askCur = 0
		m.askSels = make([]int, len(req.Questions))
		return m, nil

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
				if m.cancelStream != nil {
					m.cancelStream()
					m.cancelStream = nil
				}
				m.session.FinishAssistantMessage()
				if last := m.session.LastMessage(); last != nil {
					last.Error = "interrupted"
				}
				m.streaming = streamIdle
				m.chat.Rebuild(m.session.Messages)
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
		ctx, cancel := context.WithCancel(context.Background())
		m.cancelStream = cancel
		return m, RunAgent(ctx, m.Program, m.agent, m.driver, m.session, m.registry, m.cfg.Model)

	case streamDeltaMsg:
		wasAtBottom := m.chat.AtBottom()
		m.session.AppendDelta(string(msg))
		if wasAtBottom {
			m.chat.RebuildAndScroll(m.session.Messages)
		} else {
			m.chat.Rebuild(m.session.Messages)
		}
		return m, nil

	case thinkingStartMsg:
		m.session.SetThinkingSignature(msg.Signature)
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

	case toolCallMsg:
		if msg.IsDone {
			m.chat.RebuildAndScroll(m.session.Messages)
		} else {
			m.chat.RebuildAndScroll(m.session.Messages)
		}
		return m, nil

	case streamDoneMsg:
		if m.streaming != streamRunning {
			return m, nil
		}
		m.streaming = streamIdle
		m.chat.RebuildAndScroll(m.session.Messages)
		return m, nil

	case turnCompleteMsg:
		if m.streaming == streamRunning {
			m.chat.RebuildAndScroll(m.session.Messages)
		}
		return m, nil

	case usageMsg:
		m.inputTokens = msg.InputTokens
		m.outputTokens = msg.OutputTokens
		return m, nil

	case streamErrMsg:
		if m.streaming != streamRunning {
			return m, nil
		}
		m.session.FinishAssistantMessage()
		if last := m.session.LastMessage(); last != nil {
			last.Error = msg.Err.Error()
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

	if m.streaming != streamRunning {
		if key, ok := msg.(tea.KeyPressMsg); ok {
			if key.String() == "ctrl+t" {
				m.toggleLastThinking()
				m.chat.Rebuild(m.session.Messages)
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) updateAskMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		req := m.askReq
		q := req.Questions[m.askQ]
		key := msg.String()

		switch key {
		case "esc":
			resp := askuser.Response{Answers: make(map[string]string)}
			for i, qi := range req.Questions {
				if m.askSels[i] < len(qi.Options) {
					resp.Answers[qi.Header] = qi.Options[m.askSels[i]].Label
				} else if len(qi.Options) > 0 {
					resp.Answers[qi.Header] = qi.Options[0].Label
				}
			}
			req.ResponseCh <- resp
			m.askReq = nil
			m.chat.Rebuild(m.session.Messages)
			return m, nil

		case "up", "k":
			if m.askCur > 0 {
				m.askCur--
			}
			return m, nil

		case "down", "j":
			if m.askCur < len(q.Options)-1 {
				m.askCur++
			}
			return m, nil

		case "enter":
			m.askSels[m.askQ] = m.askCur
			if m.askQ < len(req.Questions)-1 {
				m.askQ++
				m.askCur = m.askSels[m.askQ]
			} else {
				resp := askuser.Response{Answers: make(map[string]string)}
				for i, qi := range req.Questions {
					if m.askSels[i] < len(qi.Options) {
						resp.Answers[qi.Header] = qi.Options[m.askSels[i]].Label
					}
				}
				req.ResponseCh <- resp
				m.askReq = nil
				m.chat.Rebuild(m.session.Messages)
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *Model) View() tea.View {
	if m.quitting {
		return tea.NewView("Goodbye!\n")
	}

	workDir, _ := os.Getwd()
	header := renderHeader(m.cfg.Model, workDir, m.width)
	chatView := m.chat.View()

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

	var bottom string
	if m.askReq != nil {
		bottom = m.renderAskForm()
	} else {
		bottom = m.input.View()
	}

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		body,
		"",
		bottom,
	))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m *Model) renderAskForm() string {
	req := m.askReq
	if req == nil || m.askQ >= len(req.Questions) {
		return ""
	}
	q := req.Questions[m.askQ]

	sel := lipgloss.NewStyle().Foreground(accent).Bold(true)
	unsel := lipgloss.NewStyle().Foreground(mutedFg)
	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("#D1D5DB"))

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(prompt.Render("  " + q.Text))
	sb.WriteString("\n\n")

	for i, opt := range q.Options {
		var prefix string
		if i == m.askCur {
			prefix = "  ▶ "
		} else {
			prefix = "    "
		}
		label := prefix + opt.Label
		if opt.Description != "" {
			label += "  — " + opt.Description
		}
		if i == m.askCur {
			sb.WriteString(sel.Render(label))
		} else {
			sb.WriteString(unsel.Render(label))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(prompt.Render("  ↑↓ navigate  Enter select  Esc cancel"))
	return sb.String()
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

	tcID, ok := m.chat.ToolCallLineToID(contentLine)
	if ok {
		m.chat.toolCallExpanded[tcID] = !m.chat.toolCallExpanded[tcID]
		m.chat.Rebuild(m.session.Messages)
		return
	}

	msgIdx, ok := m.chat.ThinkingLineToMsg(contentLine)
	if !ok {
		return
	}

	if msgIdx < len(m.session.Messages) {
		m.session.Messages[msgIdx].ThinkingExpanded = !m.session.Messages[msgIdx].ThinkingExpanded
		m.chat.Rebuild(m.session.Messages)
	}
}

func (m *Model) toggleLastThinking() {
	for i := len(m.session.Messages) - 1; i >= 0; i-- {
		if m.session.Messages[i].Thinking != "" {
			m.session.Messages[i].ThinkingExpanded = !m.session.Messages[i].ThinkingExpanded
			return
		}
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
