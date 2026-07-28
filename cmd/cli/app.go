package main

import (
	"os"
	"strings"
	"time"

	"github.com/cylixlee/tux/app"
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/open-portfolios/codefolio/cmd/cli/controller"
	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
	"github.com/open-portfolios/codefolio/internal/svc"
)

type App struct {
	controller      *controller.Controller
	workDir         string
	editor          *state.State[builtin.TextareaState]
	viewport        *state.State[builtin.ViewportState]
	spinner         *state.State[int]
	askOpen         *state.State[bool]
	handleEditorKey func(input.KeyEvent) bool
	moveAsk         func(int)
	confirmAsk      func()
	respondDefaults func()
}

func NewApp(cfg *conf.Global, agent *svc.Agent, session domain.Session, promptService domain.PromptService, envService domain.EnvironmentService, askUserCh chan askuser.Request) *App {
	workDir, _ := os.Getwd()
	env := envService.Detect(workDir)
	env.Model = cfg.Model
	session.AddSystemMessage(promptService.BuildSystemPrompt(env))
	a := &App{
		controller: controller.New(cfg, agent, session, askUserCh),
		workDir:    shortPath(workDir),
		editor:     state.New(builtin.TextareaState{PreferredColumn: -1}),
		viewport:   state.New(builtin.ViewportState{FollowEnd: true}),
		spinner:    state.New(0),
		askOpen:    state.New(false),
	}
	a.handleEditorKey = a.editorKey
	a.moveAsk = a.controller.MoveAsk
	a.confirmAsk = a.controller.ConfirmAsk
	a.respondDefaults = func() { a.controller.RespondAsk(true) }
	return a
}

func (a *App) AttachApp(runtime *app.App) {
	a.controller.Attach(runtime, func() { a.spinner.Set(a.spinner.Value()) }, a.askOpen.Set)
	runtime.OnTimer(100*time.Millisecond, func() {
		if a.controller.Running() {
			a.spinner.Set((a.spinner.Value() + 1) % 4)
		}
	})
}

func (a *App) KeyMap() input.KeyMap {
	return input.KeyMap{input.OnStop(input.Ctrl('t'), func(input.KeyEvent) {
		if !a.controller.Running() {
			a.controller.ToggleLastThinking()
		}
	})}
}

func (a *App) editorKey(event input.KeyEvent) bool {
	if a.controller.Cancelling() {
		return true
	}
	if a.controller.Running() {
		if event.Key == input.KeyEscape {
			a.controller.Cancel()
		}
		return true
	}
	if event.Rune == 'c' && event.Modifiers == input.ModCtrl {
		if a.editor.Value().Value == "" {
			return false
		}
		a.editor.Set(builtin.TextareaState{PreferredColumn: -1, CursorOn: true})
		return true
	}
	if event.Key == input.KeyEnter && event.Modifiers == 0 {
		content := strings.TrimSpace(a.editor.Value().Value)
		if content == "" {
			return false
		}
		state := a.editor.Value()
		a.controller.Start(content, &state)
		a.editor.Set(state)
		a.viewport.Set(builtin.ViewportState{FollowEnd: true})
		return true
	}
	if event.Modifiers == input.ModCtrl && event.Key == input.KeyArrowUp {
		a.setEditor(a.controller.HistoryUp(a.editor.Value().Value))
		return true
	}
	if event.Modifiers == input.ModCtrl && event.Key == input.KeyArrowDown {
		a.setEditor(a.controller.HistoryDown())
		return true
	}
	return false
}

func (a *App) setEditor(value string) {
	a.editor.Set(builtin.TextareaState{Value: value, Cursor: len(value), PreferredColumn: -1, CursorOn: true})
}

func shortPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

var _ renderer.Component = (*App)(nil)
