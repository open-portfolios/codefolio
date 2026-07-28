package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cylixlee/tux/app"
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
	"github.com/mattn/go-runewidth"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
	"github.com/open-portfolios/codefolio/internal/svc"
)

const (
	sidebarWidth = 28
	inputMaxH    = 8
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

type streamState uint8

const (
	streamIdle streamState = iota
	streamRunning
)

type toolView struct {
	ID       string
	Name     string
	Input    string
	Output   string
	Done     bool
	IsError  bool
	Elapsed  time.Duration
	Expanded bool
}

type messageView struct {
	ID               string
	Role             string
	Content          string
	Thinking         string
	ThinkingExpanded bool
	Streaming        bool
	Error            string
	Tools            []*toolView
}

// Model owns all presentation state. Agent goroutines only submit domain
// events through App.Dispatch; they never render from or mutate this state.
type Model struct {
	cfg     *conf.Global
	agent   *svc.Agent
	session domain.Session

	app       *app.App
	appMu     sync.RWMutex
	askUserCh <-chan askuser.Request

	messages []*messageView
	editor   *state.State[builtin.TextareaState]
	viewport *state.State[builtin.ViewportState]
	spinner  *state.State[int]
	askOpen  *state.State[bool]

	streaming  streamState
	cancelling bool
	runID      uint64
	cancel     context.CancelFunc
	history    []string
	historyAt  int
	draft      string

	chatWidth int
	regions   []clickRegion

	askReq *askuser.Request
	askQ   int
	askCur int
	askSel []int

	inputTokens  int64
	outputTokens int64
}

type clickRegion struct {
	rect   renderer.Rect
	kind   string
	msg    int
	toolID string
}

func NewModel(cfg *conf.Global, agent *svc.Agent, session domain.Session, promptService domain.PromptService, envService domain.EnvironmentService, askUserCh chan askuser.Request) *Model {
	workDir, _ := os.Getwd()
	env := envService.Detect(workDir)
	env.Model = cfg.Model
	session.AddSystemMessage(promptService.BuildSystemPrompt(env))

	return &Model{
		cfg:       cfg,
		agent:     agent,
		session:   session,
		askUserCh: askUserCh,
		editor:    state.New(builtin.TextareaState{PreferredColumn: -1}),
		viewport:  state.New(builtin.ViewportState{FollowEnd: true}),
		spinner:   state.New(0),
		askOpen:   state.New(false),
		historyAt: -1,
		chatWidth: 80,
	}
}

func (m *Model) AttachApp(runtime *app.App) {
	m.appMu.Lock()
	m.app = runtime
	m.appMu.Unlock()

	go m.askUserLoop()
	runtime.OnTimer(100*time.Millisecond, func() {
		if m.streaming == streamRunning {
			m.spinner.Set((m.spinner.Value() + 1) % len(spinner))
		}
	})
}

func (m *Model) runtime() *app.App {
	m.appMu.RLock()
	defer m.appMu.RUnlock()
	return m.app
}

func (m *Model) askUserLoop() {
	for req := range m.askUserCh {
		runtime := m.runtime()
		if runtime == nil {
			continue
		}
		_ = runtime.Dispatch(context.Background(), func() {
			m.askReq = &req
			m.askQ, m.askCur = 0, 0
			m.askSel = make([]int, len(req.Questions))
			m.askOpen.Set(true)
		})
	}
}

func (m *Model) KeyMap() input.KeyMap {
	return input.KeyMap{
		input.OnStop(input.Ctrl('t'), func(input.KeyEvent) {
			if m.streaming != streamRunning {
				m.toggleLastThinking()
			}
		}),
	}
}

func (m *Model) Render(ctx renderer.Context) *renderer.Element {
	screenWidth, _ := ctx.Size()
	if screenWidth <= 0 {
		screenWidth = 80
	}
	workDir, _ := os.Getwd()
	header := builtin.CreateTextBlock(ctx, builtin.TextBlockProps{
		Key: "header", Width: max(m.chatWidth, 40), Text: fmt.Sprintf("Codefolio  %s  %s", m.cfg.Model, shortPath(workDir)),
		Fg: primary, Style: style.Bold,
	})

	transcript := (&transcriptComponent{model: m}).Render(ctx)
	chat := builtin.CreateViewport(ctx, builtin.ViewportProps{Key: "chat", State: m.viewport, ScrollY: true}, transcript)
	chat.SetFlex(1)

	mainChildren := []renderer.Component{chat}
	if screenWidth >= 120 {
		mainChildren = append(mainChildren, m.sidebar(ctx))
	}
	main := builtin.CreateRow(ctx, builtin.RowProps{Key: "main"}, mainChildren...)
	main.SetFlex(1)

	inputHeight := builtin.TextareaPreferredHeight(m.editor.Value().Value, max(m.chatWidth, 20), 1, inputMaxH)
	editor := builtin.CreateTextarea(ctx, builtin.TextareaProps{
		Key: "editor", State: m.editor, Placeholder: "Ask something...", MinHeight: 1, MaxHeight: inputMaxH,
		Fg: primary, OnKey: m.handleEditorKey,
	})
	editor.Focus()
	editor.SetLayoutHeight(inputHeight)
	help := builtin.CreateTextBlock(ctx, builtin.TextBlockProps{
		Key: "help", Width: max(m.chatWidth, 40), Text: "Enter send | modified Enter newline | Esc cancel | Ctrl+Up/Down history | Ctrl+C clear/quit",
		Fg: muted, Style: style.Dim,
	})
	bottom := builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "bottom", Padding: 1}, editor, help)

	children := []renderer.Component{header, main, bottom, m.askModal(ctx)}
	return builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "root", Padding: 1, Gap: 1}, children...)
}

func (m *Model) sidebar(ctx renderer.Context) renderer.Component {
	text := fmt.Sprintf("SESSION\nMessages  %d\nInput     %d\nOutput    %d\nStatus    %s", len(m.messages), m.inputTokens, m.outputTokens, m.statusText())
	return builtin.CreateBorder(ctx, builtin.BorderProps{Key: "sidebar", Title: " Session ", BorderType: builtin.BorderSingle, BorderColor: muted},
		builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "sidebar-content", Text: text, Width: sidebarWidth - 4, Fg: muted}))
}

func (m *Model) statusText() string {
	if m.cancelling {
		return "cancelling"
	}
	if m.streaming == streamRunning {
		return "working"
	}
	return "idle"
}

func (m *Model) handleEditorKey(ev input.KeyEvent) bool {
	if m.cancelling {
		return true
	}
	if m.streaming == streamRunning {
		if ev.Key == input.KeyEscape && m.cancel != nil {
			m.cancel()
			m.cancel = nil
			m.finishCurrent("interrupted")
			m.cancelling = true
		}
		return true
	}

	if ev.Rune == 'c' && ev.Modifiers == input.ModCtrl {
		if m.editor.Value().Value != "" {
			m.editor.Set(builtin.TextareaState{PreferredColumn: -1, CursorOn: true})
			return true
		}
		return false
	}
	if ev.Key == input.KeyEnter && ev.Modifiers == 0 {
		content := strings.TrimSpace(m.editor.Value().Value)
		if content == "" {
			return false
		}
		m.startRun(content)
		return true
	}
	if ev.Modifiers == input.ModCtrl && ev.Key == input.KeyArrowUp {
		m.historyUp()
		return true
	}
	if ev.Modifiers == input.ModCtrl && ev.Key == input.KeyArrowDown {
		m.historyDown()
		return true
	}
	return false
}

func (m *Model) startRun(content string) {
	m.history = append(m.history, content)
	m.historyAt = len(m.history)
	m.draft = ""
	m.editor.Set(builtin.TextareaState{PreferredColumn: -1, CursorOn: true})
	m.session.AddUserMessage(content)
	m.session.StartAssistantMessage()
	m.messages = append(m.messages, &messageView{ID: fmt.Sprintf("user-%d", len(m.messages)+1), Role: "user", Content: content})
	m.messages = append(m.messages, &messageView{ID: fmt.Sprintf("assistant-%d", len(m.messages)+1), Role: "assistant", Streaming: true})
	m.streaming = streamRunning
	m.runID++
	runID := m.runID
	m.viewport.Set(builtin.ViewportState{FollowEnd: true})
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	runtime := m.runtime()
	go func() {
		visitor := &eventVisitor{model: m, runtime: runtime, runID: runID, ctx: ctx}
		defer func() {
			// The stream context may already be cancelled, but this lifecycle
			// cleanup must still reach the UI loop so a new request can start.
			m.session.FinishAssistantMessage()
			if runtime == nil {
				return
			}
			_ = runtime.Dispatch(context.Background(), func() {
				if runID == m.runID {
					m.cancelling = false
				}
			})
		}()
		if err := m.agent.Run(ctx, m.session, m.cfg, visitor); err != nil {
			_ = visitor.post(func() { m.applyError(runID, err) })
		}
	}()
}

func (m *Model) historyUp() {
	if len(m.history) == 0 || m.historyAt <= 0 {
		return
	}
	if m.historyAt == len(m.history) {
		m.draft = m.editor.Value().Value
	}
	m.historyAt--
	m.setEditor(m.history[m.historyAt])
}

func (m *Model) historyDown() {
	if m.historyAt >= len(m.history) {
		return
	}
	m.historyAt++
	if m.historyAt == len(m.history) {
		m.setEditor(m.draft)
		return
	}
	m.setEditor(m.history[m.historyAt])
}

func (m *Model) setEditor(value string) {
	m.editor.Set(builtin.TextareaState{Value: value, Cursor: len(value), PreferredColumn: -1, CursorOn: true})
}

func (m *Model) currentAssistant() *messageView {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "assistant" {
			return m.messages[i]
		}
	}
	return nil
}

func (m *Model) finishCurrent(reason string) {
	if m.streaming != streamRunning {
		return
	}
	if msg := m.currentAssistant(); msg != nil {
		msg.Streaming = false
		if reason != "" {
			msg.Error = reason
		}
	}
	m.streaming = streamIdle
	m.cancel = nil
}

func (m *Model) applyError(runID uint64, err error) {
	if runID != m.runID || m.streaming != streamRunning {
		return
	}
	m.finishCurrent(err.Error())
}

func (m *Model) toggleLastThinking() {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Thinking != "" {
			m.messages[i].ThinkingExpanded = !m.messages[i].ThinkingExpanded
			return
		}
	}
}

func (m *Model) askModal(ctx renderer.Context) renderer.Component {
	question := "Waiting for a question"
	items := []string(nil)
	if m.askReq != nil && m.askQ < len(m.askReq.Questions) {
		q := m.askReq.Questions[m.askQ]
		question = q.Text
		items = make([]string, len(q.Options))
		for i, option := range q.Options {
			items[i] = option.Label
			if option.Description != "" {
				items[i] += " - " + option.Description
			}
		}
	}
	keys := builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "ask-keys", Text: question, Width: max(m.chatWidth/2, 30), Fg: primary})
	keys.SetHandleKeyFn(func(ev input.KeyEvent) bool { return m.handleAskKey(ev, ctx) })
	form := builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "ask-form", Padding: 1, Gap: 1},
		keys,
		builtin.CreateList(ctx, builtin.ListProps{Key: "ask-options", Items: items, Selected: m.askCur, Height: min(max(len(items), 1), 4), Fg: primary}),
		builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "ask-help", Text: "Up/Down select | Enter confirm | Esc use defaults", Width: max(m.chatWidth/2, 30), Fg: muted, Style: style.Dim}),
	)
	return builtin.CreateModal(ctx, builtin.ModalProps{Key: "ask-modal", Open: m.askOpen, Backdrop: "dim", CloseOnEscape: false, Title: " Agent question ", Border: builtin.BorderRounded, BorderColor: accent}, form)
}

func (m *Model) handleAskKey(ev input.KeyEvent, ctx renderer.Context) bool {
	if m.askReq == nil || m.askQ >= len(m.askReq.Questions) {
		return false
	}
	q := m.askReq.Questions[m.askQ]
	switch {
	case ev.Key == input.KeyArrowUp || ev.Rune == 'k':
		m.askCur = max(m.askCur-1, 0)
	case ev.Key == input.KeyArrowDown || ev.Rune == 'j':
		m.askCur = min(m.askCur+1, len(q.Options)-1)
	case ev.Key == input.KeyEnter:
		m.askSel[m.askQ] = m.askCur
		if m.askQ < len(m.askReq.Questions)-1 {
			m.askQ++
			m.askCur = m.askSel[m.askQ]
		} else {
			m.respondAsk(false)
		}
	case ev.Key == input.KeyEscape:
		m.respondAsk(true)
	default:
		return false
	}
	ctx.MarkDirty()
	return true
}

func (m *Model) respondAsk(defaults bool) {
	if m.askReq == nil {
		return
	}
	response := askuser.Response{Answers: make(map[string]string)}
	for i, q := range m.askReq.Questions {
		selected := m.askSel[i]
		if defaults {
			selected = 0
		}
		if selected >= 0 && selected < len(q.Options) {
			response.Answers[q.Header] = q.Options[selected].Label
		}
	}
	select {
	case m.askReq.ResponseCh <- response:
	default:
	}
	m.askReq = nil
	m.askOpen.Set(false)
}

type eventVisitor struct {
	domain.BaseEventVisitor
	model   *Model
	runtime *app.App
	runID   uint64
	ctx     context.Context
}

func (v *eventVisitor) post(fn func()) error {
	if v.runtime == nil {
		return app.ErrStopped
	}
	return v.runtime.Dispatch(v.ctx, fn)
}

func (v *eventVisitor) VisitStream(e domain.StreamEvent) error {
	return v.post(func() {
		if v.runID == v.model.runID && v.model.streaming == streamRunning {
			if m := v.model.currentAssistant(); m != nil {
				m.Content += e.Content
			}
		}
	})
}

func (v *eventVisitor) VisitThink(e domain.ThinkEvent) error {
	return v.post(func() {
		if v.runID == v.model.runID && v.model.streaming == streamRunning {
			if m := v.model.currentAssistant(); m != nil {
				m.Thinking += e.Content
			}
		}
	})
}

func (v *eventVisitor) VisitThinkStart(domain.ThinkStartEvent) error { return nil }

func (v *eventVisitor) VisitToolCall(e domain.ToolCallEvent) error {
	return v.post(func() {
		if v.runID == v.model.runID && v.model.streaming == streamRunning {
			if m := v.model.currentAssistant(); m != nil {
				m.Tools = append(m.Tools, &toolView{ID: e.ID, Name: e.Name, Input: e.Input})
			}
		}
	})
}

func (v *eventVisitor) VisitToolResult(e domain.ToolResultEvent) error {
	return v.post(func() {
		if v.runID != v.model.runID || v.model.streaming != streamRunning {
			return
		}
		if m := v.model.currentAssistant(); m != nil {
			for _, tool := range m.Tools {
				if tool.ID == e.ID {
					tool.Done, tool.Output, tool.IsError, tool.Elapsed = true, e.Output, e.IsError, e.Elapsed
					break
				}
			}
		}
	})
}

func (v *eventVisitor) VisitTurnComplete(domain.TurnCompleteEvent) error { return nil }

func (v *eventVisitor) VisitLoopComplete(domain.LoopCompleteEvent) error {
	return v.post(func() {
		if v.runID == v.model.runID {
			v.model.finishCurrent("")
		}
	})
}

func (v *eventVisitor) VisitUsage(e domain.UsageEvent) error {
	return v.post(func() {
		if v.runID == v.model.runID {
			v.model.inputTokens, v.model.outputTokens = e.InputTokens, e.OutputTokens
		}
	})
}

func (v *eventVisitor) VisitError(e domain.ErrorEvent) error {
	return v.post(func() { v.model.applyError(v.runID, e.Err) })
}

type transcriptComponent struct{ model *Model }

func (t *transcriptComponent) Render(ctx renderer.Context) *renderer.Element {
	e := &renderer.Element{}
	e.SetKey("transcript")
	e.SetTag("transcript")
	e.SetLayoutWidth(0)
	e.SetLayoutHeight(t.model.transcriptHeight(max(t.model.chatWidth, 20)))
	e.SetPaintFn(func(draw renderer.DrawContext, box renderer.Rect) {
		width := max(box.Width, 20)
		if width != t.model.chatWidth {
			t.model.chatWidth = width
			ctx.MarkDirty()
		}
		t.model.drawTranscript(draw, box, e)
	})
	e.SetHandleMouseFn(func(ev input.KeyEvent) bool {
		if ev.Mouse.Action != input.MousePress || ev.Mouse.Button != input.MouseLeft {
			return false
		}
		for _, region := range t.model.regions {
			if contains(region.rect, ev.Mouse.X, ev.Mouse.Y) {
				switch region.kind {
				case "thinking":
					t.model.messages[region.msg].ThinkingExpanded = !t.model.messages[region.msg].ThinkingExpanded
				case "tool":
					for _, tool := range t.model.messages[region.msg].Tools {
						if tool.ID == region.toolID {
							tool.Expanded = !tool.Expanded
						}
					}
				}
				ctx.MarkDirty()
				return true
			}
		}
		return false
	})
	return e
}

func (m *Model) transcriptHeight(width int) int { return len(m.transcriptLines(width, false)) }

type visualLine struct {
	text   string
	fg     style.Color
	bg     style.Color
	attrs  style.Attr
	kind   string
	msg    int
	toolID string
}

func (m *Model) transcriptLines(width int, regions bool) []visualLine {
	width = max(width, 20)
	var lines []visualLine
	for msgIndex, msg := range m.messages {
		switch msg.Role {
		case "user":
			lines = append(lines, wrapVisual("| "+msg.Content, width, userBar, 0, "", msgIndex, "")...)
		case "assistant":
			if msg.Thinking != "" {
				prefix := "+ Thinking"
				if msg.Streaming && msg.Content == "" {
					prefix = spinner[m.spinner.Value()%len(spinner)] + " Thinking"
				} else if msg.ThinkingExpanded {
					prefix = "- Thinking"
				}
				lines = append(lines, visualLine{text: prefix, fg: muted, attrs: style.Italic, kind: "thinking", msg: msgIndex})
				if msg.ThinkingExpanded {
					lines = append(lines, wrapVisual("  "+msg.Thinking, width, muted, style.Italic, "", msgIndex, "")...)
				}
			}
			if msg.Content != "" {
				lines = append(lines, markdownVisual(msg.Content, width, msgIndex)...)
			}
			if msg.Error != "" {
				lines = append(lines, wrapVisual("Error: "+msg.Error, width, errorFg, 0, "", msgIndex, "")...)
			}
			for _, tool := range msg.Tools {
				label := toolLabel(tool)
				if !tool.Done {
					label = spinner[m.spinner.Value()%len(spinner)] + " " + label
				}
				if tool.Done && showToolOutput(tool.Name) {
					if tool.Expanded {
						label = "- " + label
					} else {
						label = "+ " + label
					}
				}
				if tool.Done && tool.Elapsed > 0 {
					label += " (" + tool.Elapsed.Round(time.Millisecond).String() + ")"
				}
				kind := ""
				if tool.Done && showToolOutput(tool.Name) {
					kind = "tool"
				}
				lines = append(lines, wrapVisual("  "+label, width, muted, 0, kind, msgIndex, tool.ID)...)
				if tool.Done && tool.Expanded && tool.Output != "" {
					fg := muted
					if tool.IsError {
						fg = errorFg
					}
					lines = append(lines, wrapVisual("    "+tool.Output, width, fg, 0, "", msgIndex, "")...)
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

func (m *Model) drawTranscript(draw renderer.DrawContext, box renderer.Rect, e *renderer.Element) {
	lines := m.transcriptLines(max(box.Width, 20), true)
	m.regions = m.regions[:0]
	for row, line := range lines {
		if row >= box.Height {
			break
		}
		if line.bg != (style.Color{}) {
			draw.Fill(renderer.Rect{X: box.X, Y: box.Y + row, Width: box.Width, Height: 1}, renderer.Cell{Rune: ' ', Style: style.Style{Bg: line.bg}})
		}
		draw.WriteStringWide(box.X, box.Y+row, line.text, style.Style{Fg: line.fg, Bg: line.bg, Attrs: line.attrs})
		if line.kind != "" {
			m.regions = append(m.regions, clickRegion{rect: renderer.Rect{X: draw.Origin.X + box.X, Y: draw.Origin.Y + box.Y + row, Width: box.Width, Height: 1}, kind: line.kind, msg: line.msg, toolID: line.toolID})
		}
	}
	e.SetRect(renderer.Rect{X: draw.Origin.X + box.X, Y: draw.Origin.Y + box.Y, Width: box.Width, Height: len(lines)})
}

func wrapVisual(text string, width int, fg style.Color, attrs style.Attr, kind string, msg int, toolID string) []visualLine {
	var result []visualLine
	for source := range strings.SplitSeq(text, "\n") {
		if source == "" {
			result = append(result, visualLine{fg: fg, attrs: attrs, kind: kind, msg: msg, toolID: toolID})
			continue
		}
		var line strings.Builder
		lineWidth := 0
		for _, r := range source {
			rw := max(runewidth.RuneWidth(r), 1)
			if lineWidth+rw > width && line.Len() > 0 {
				result = append(result, visualLine{text: line.String(), fg: fg, attrs: attrs, kind: kind, msg: msg, toolID: toolID})
				line.Reset()
				lineWidth = 0
			}
			line.WriteRune(r)
			lineWidth += rw
		}
		result = append(result, visualLine{text: line.String(), fg: fg, attrs: attrs, kind: kind, msg: msg, toolID: toolID})
	}
	return result
}

// markdownVisual produces styled terminal cells instead of ANSI strings. It
// covers the blocks most common in agent output while keeping the transcript
// viewport's width-aware line model intact.
func markdownVisual(content string, width, msg int) []visualLine {
	var result []visualLine
	inCode := false
	for source := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(source)
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			if inCode && len(trimmed) > 3 {
				result = append(result, visualLine{text: "  " + strings.TrimSpace(trimmed[3:]), fg: muted, bg: codeBg, attrs: style.Dim, msg: msg})
			}
			continue
		}
		if inCode {
			for _, line := range wrapVisual("  "+source, width, codeFg, 0, "", msg, "") {
				line.bg = codeBg
				result = append(result, line)
			}
			continue
		}

		fg, attrs := primary, style.Attr(0)
		text := source
		switch {
		case strings.HasPrefix(trimmed, "### "):
			text, fg, attrs = strings.TrimPrefix(trimmed, "### "), accent, style.Bold
		case strings.HasPrefix(trimmed, "## "):
			text, fg, attrs = strings.TrimPrefix(trimmed, "## "), accent, style.Bold
		case strings.HasPrefix(trimmed, "# "):
			text, fg, attrs = strings.TrimPrefix(trimmed, "# "), accent, style.Bold
		case strings.HasPrefix(trimmed, "> "):
			text, fg, attrs = "| "+strings.TrimPrefix(trimmed, "> "), muted, style.Italic
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			text = "• " + trimmed[2:]
		default:
			text = inlineCodeMarker(source)
		}
		result = append(result, wrapVisual(text, width, fg, attrs, "", msg, "")...)
	}
	return result
}

// Inline code receives a visible marker without injecting terminal ANSI
// escapes. Span-aware inline styling can later replace this small adapter.
func inlineCodeMarker(text string) string {
	if strings.Count(text, "`") < 2 {
		return text
	}
	return strings.ReplaceAll(text, "`", "'")
}

func toolLabel(tool *toolView) string {
	verb := map[string]string{"ReadFile": "Read", "WriteFile": "Wrote", "EditFile": "Edited", "Glob": "Glob", "Grep": "Grep", "Bash": "Bash", "AskUserQuestion": "Question"}[tool.Name]
	if verb == "" {
		verb = tool.Name
	}
	field := "file_path"
	if tool.Name == "Glob" {
		field = "pattern"
	}
	if tool.Name == "Grep" {
		field = "pattern"
	}
	if tool.Name == "Bash" {
		field = "command"
	}
	value := extractField(tool.Input, field)
	if value == "" {
		return verb
	}
	return verb + " " + value
}

func showToolOutput(name string) bool { return name == "Bash" || name == "AskUserQuestion" }

func extractField(raw, key string) string {
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

func shortPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
