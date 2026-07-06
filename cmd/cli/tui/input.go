package tui

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type InputModel struct {
	textarea textarea.Model
	history  []string
	histIdx  int
	width    int
}

func NewInputModel(w int) InputModel {
	ta := textarea.New()
	ta.Prompt = ""
	ta.Placeholder = "Ask something..."
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.MaxHeight = 6
	ta.SetWidth(w)

	styles := ta.Styles()
	styles.Focused.CursorLine = lipgloss.NewStyle()
	styles.Blurred.CursorLine = lipgloss.NewStyle()
	styles.Cursor.Blink = false
	ta.SetStyles(styles)

	ta.Focus()

	return InputModel{
		textarea: ta,
		history:  make([]string, 0),
		histIdx:  -1,
		width:    w,
	}
}

func (m *InputModel) SetSize(w int) {
	m.width = w
	m.textarea.SetWidth(w)
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) View() string {
	inputView := m.textarea.View()

	helpText := "enter · send  |  ctrl+enter · newline  |  ctrl+↑↓ · history  |  ctrl+c · quit"
	help := helpBarStyle.Render(helpText)

	return lipgloss.JoinVertical(lipgloss.Left, inputView, help)
}

type inputSentMsg struct {
	content string
}

func (m *InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return *m, tea.Quit

		case "enter":
			content := strings.TrimSpace(m.textarea.Value())
			if content == "" {
				m.textarea, _ = m.textarea.Update(msg)
				return *m, nil
			}
			m.history = append(m.history, content)
			m.histIdx = len(m.history)
			m.textarea.Reset()
			return *m, func() tea.Msg {
				return inputSentMsg{content: content}
			}

		case "ctrl+up":
			if m.histIdx > 0 {
				m.histIdx--
				m.textarea.SetValue(m.history[m.histIdx])
			}
			return *m, nil

		case "ctrl+down":
			if m.histIdx < len(m.history)-1 {
				m.histIdx++
				m.textarea.SetValue(m.history[m.histIdx])
			} else {
				m.histIdx = len(m.history)
				m.textarea.SetValue("")
			}
			return *m, nil

		case "ctrl+enter":
			m.textarea.InsertString("\n")
			return *m, nil
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return *m, cmd
}

var helpBarStyle = lipgloss.NewStyle().
	Foreground(mutedFg).
	Padding(0, 1)
