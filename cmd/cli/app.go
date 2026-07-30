package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cylixlee/tux/app"
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/open-portfolios/codefolio/cmd/cli/components"
	"github.com/open-portfolios/codefolio/cmd/cli/controller"
	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/approval"
	"github.com/open-portfolios/codefolio/internal/infra/mcp"
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
	approvalOpen    *state.State[bool]
	mcpStatus       *state.State[components.MCPStatus]
	mcpManager      *mcp.Manager
	toolRegistrar   domain.ToolRegistrar
	handleEditorKey func(input.KeyEvent) bool
	moveAsk         func(int)
	confirmAsk      func()
	respondDefaults func()
	approveOnce     func()
	approveSession  func()
	denyApproval    func()
	composerEnabled bool
	composerFocus   bool
}

func NewApp(cfg *conf.Struct, agent *svc.Agent, session domain.Session, promptService domain.PromptService, envService domain.EnvironmentService, mcpManager *mcp.Manager, toolRegistrar domain.ToolRegistrar, askUserCh chan askuser.Request, approvalCh chan *approval.Request) *App {
	workDir, _ := os.Getwd()
	agent.WorkDir = workDir
	env := envService.Detect(workDir)
	env.Model = cfg.Model
	session.AddSystemMessage(promptService.BuildSystemPrompt(env))
	a := &App{
		controller:    controller.New(cfg, agent, session, askUserCh, approvalCh),
		workDir:       shortPath(workDir),
		editor:        state.New(builtin.TextareaState{PreferredColumn: -1}),
		viewport:      state.New(builtin.ViewportState{FollowEnd: true}),
		spinner:       state.New(0),
		askOpen:       state.New(false),
		approvalOpen:  state.New(false),
		mcpStatus:     state.New(mcpStatusFor(mcpManager.Summary(), false)),
		mcpManager:    mcpManager,
		toolRegistrar: toolRegistrar,
	}
	a.handleEditorKey = a.editorKey
	a.moveAsk = a.controller.MoveAsk
	a.confirmAsk = a.controller.ConfirmAsk
	a.respondDefaults = func() { a.controller.RespondAsk(true) }
	a.approveOnce = a.controller.ApproveOnce
	a.approveSession = a.controller.ApproveSession
	a.denyApproval = a.controller.DenyApproval
	return a
}

func (a *App) AttachApp(runtime *app.App) {
	a.controller.Attach(runtime, func() { a.spinner.Set(a.spinner.Value()) }, a.askOpen.Set, a.approvalOpen.Set)
	runtime.OnTimer(100*time.Millisecond, func() {
		if a.controller.Running() {
			a.spinner.Set((a.spinner.Value() + 1) % 4)
		}
	})
	a.startMCP(runtime)
}

func (a *App) startMCP(runtime *app.App) {
	if a.mcpManager == nil || a.toolRegistrar == nil {
		return
	}
	summary := a.mcpManager.Summary()
	if summary.Configured == 0 {
		return
	}
	a.mcpStatus.Set(mcpStatusFor(summary, true))
	go func() {
		tools, err := a.mcpManager.Discover(context.Background())
		if err == nil {
			for _, tool := range tools {
				if registerErr := a.toolRegistrar.Register(tool); registerErr != nil {
					err = fmt.Errorf("register MCP tool: %w", registerErr)
					break
				}
			}
		}
		_ = runtime.Dispatch(context.Background(), func() {
			a.mcpStatus.Set(mcpStatusFor(a.mcpManager.Summary(), false))
		})
	}()
}

func (a *App) MCPStatus(ctx renderer.Context) components.MCPStatus {
	if a.mcpStatus == nil {
		return mcpStatusFor(mcp.Summary{}, false)
	}
	return a.mcpStatus.Get(ctx)
}

func mcpStatusFor(summary mcp.Summary, connecting bool) components.MCPStatus {
	if summary.Configured == 0 {
		return components.MCPStatus{Label: "○ No servers configured", Color: components.Theme.TextMuted}
	}
	if connecting {
		return components.MCPStatus{Label: fmt.Sprintf("◌ Connecting %d server(s)...", summary.Configured), Color: components.Theme.Warning}
	}
	if summary.Ready == summary.Configured {
		return components.MCPStatus{Label: fmt.Sprintf("● %d server(s) · %d tools", summary.Ready, summary.Tools), Color: components.Theme.Success}
	}
	if summary.Ready > 0 {
		return components.MCPStatus{Label: fmt.Sprintf("▲ %d ready · %d unavailable", summary.Ready, summary.Unavailable), Color: components.Theme.Warning}
	}
	return components.MCPStatus{Label: fmt.Sprintf("× %d server(s) unavailable", summary.Unavailable), Color: components.Theme.Error}
}

func (a *App) Shutdown() { a.controller.Shutdown() }

func (a *App) KeyMap() input.KeyMap {
	return input.KeyMap{input.OnStop(input.Ctrl('t'), func(input.KeyEvent) {
		if !a.controller.Running() {
			a.controller.ToggleLastThinking()
		}
	})}
}

func (a *App) ComposerDisabled() bool {
	enabled := !a.controller.Running() && !a.controller.Cancelling()
	a.composerFocus = enabled && !a.composerEnabled && a.viewport.Value().FollowEnd
	a.composerEnabled = enabled
	return !enabled
}

func (a *App) FocusComposer() bool { return a.composerFocus }

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
