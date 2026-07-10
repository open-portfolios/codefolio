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
	ta.MaxHeight = 8
	ta.SetWidth(w)
	ta.SetHeight(8)

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

func (m *InputModel) SetSize(w, h int) {
	m.width = w
	m.textarea.SetWidth(w)
	m.textarea.SetHeight(h)
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) View() string {
	inputView := m.textarea.View()

	helpText := "enter · send  |  ctrl+enter · newline  |  esc · cancel  |  ctrl+↑↓ · history  |  ctrl+c · quit"
	help := helpBarStyle.Render(helpText)

	return lipgloss.JoinVertical(lipgloss.Left, inputView, help)
}

type inputSentMsg struct {
	content string
}

func (m *InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		k := msg.Key()
		switch {
		case k.Code == 'c' && k.Mod&tea.ModCtrl != 0:
			return *m, tea.Quit

		case k.Mod&tea.ModCtrl != 0 && k.Code == 'j':
			m.textarea.InsertString("\n")
			return *m, nil

		case k.Code == tea.KeyEnter:
			content := strings.TrimSpace(m.textarea.Value())
			if content == "" {
				m.textarea.InsertString("\n")
				return *m, nil
			}
			m.history = append(m.history, content)
			m.histIdx = len(m.history)
			m.textarea.Reset()
			return *m, func() tea.Msg {
				return inputSentMsg{content: content}
			}

		case k.Mod&tea.ModCtrl != 0 && k.Code == tea.KeyUp:
			if m.histIdx > 0 {
				m.histIdx--
				m.textarea.SetValue(m.history[m.histIdx])
			}
			return *m, nil

		case k.Mod&tea.ModCtrl != 0 && k.Code == tea.KeyDown:
			if m.histIdx < len(m.history)-1 {
				m.histIdx++
				m.textarea.SetValue(m.history[m.histIdx])
			} else {
				m.histIdx = len(m.history)
				m.textarea.SetValue("")
			}
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
