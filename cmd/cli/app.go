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
	commandinfra "github.com/open-portfolios/codefolio/internal/infra/command"
	"github.com/open-portfolios/codefolio/internal/infra/mcp"
	"github.com/open-portfolios/codefolio/internal/infra/tools"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
	"github.com/open-portfolios/codefolio/internal/svc"
)

type App struct {
	cfg                *conf.Struct
	controller         *controller.Controller
	rootDir            string
	workDir            string
	editor             *state.State[builtin.TextareaState]
	viewport           *state.State[builtin.ViewportState]
	spinner            *state.State[int]
	askOpen            *state.State[bool]
	approvalOpen       *state.State[bool]
	resumeOpen         *state.State[bool]
	mcpStatus          *state.State[components.MCPStatus]
	mcpManager         *mcp.Manager
	toolRegistrar      domain.ToolRegistrar
	toolRegistry       domain.ToolRegistry
	quit               func()
	commands           *svc.CommandRegistry
	sessions           *svc.SessionService
	contextManager     *svc.ContextManager
	memory             *svc.MemoryService
	systemPrompt       string
	handleEditorKey    func(input.KeyEvent) bool
	moveAsk            func(int)
	confirmAsk         func()
	respondDefaults    func()
	approveOnce        func()
	approveSession     func()
	denyApproval       func()
	moveResume         func(int)
	confirmResume      func()
	cancelResume       func()
	resumeSessions     []domain.SessionInfo
	resumeSelected     int
	completionPrefix   string
	completionSelected int
	composerEnabled    bool
	composerFocus      bool
}

func NewApp(cfg *conf.Struct, agent *svc.Agent, session domain.Session, promptService domain.PromptService, envService domain.EnvironmentService, mcpManager *mcp.Manager, registry *tools.Registry, commands *svc.CommandRegistry, sessions *svc.SessionService, contextManager *svc.ContextManager, memory *svc.MemoryService, askUserCh chan askuser.Request, approvalCh chan *approval.Request) *App {
	workDir, _ := os.Getwd()
	agent.WorkDir = workDir
	env := envService.Detect(workDir)
	env.Model = cfg.Model
	systemPrompt := promptService.BuildSystemPrompt(env)
	session.AddSystemMessage(systemPrompt)
	a := &App{
		cfg:            cfg,
		controller:     controller.New(cfg, agent, session, askUserCh, approvalCh),
		rootDir:        workDir,
		workDir:        shortPath(workDir),
		editor:         state.New(builtin.TextareaState{PreferredColumn: -1}),
		viewport:       state.New(builtin.ViewportState{FollowEnd: true}),
		spinner:        state.New(0),
		askOpen:        state.New(false),
		approvalOpen:   state.New(false),
		resumeOpen:     state.New(false),
		mcpStatus:      state.New(mcpStatusFor(mcpManager.Summary(), false)),
		mcpManager:     mcpManager,
		toolRegistrar:  registry,
		toolRegistry:   registry,
		commands:       commands,
		sessions:       sessions,
		contextManager: contextManager,
		memory:         memory,
		systemPrompt:   systemPrompt,
	}
	if a.commands != nil {
		loaded := commandinfra.Load(workDir)
		a.commands.ReplaceDynamic(loaded.Commands)
	}
	a.handleEditorKey = a.editorKey
	a.moveAsk = a.controller.MoveAsk
	a.confirmAsk = a.controller.ConfirmAsk
	a.respondDefaults = func() { a.controller.RespondAsk(true) }
	a.approveOnce = a.controller.ApproveOnce
	a.approveSession = a.controller.ApproveSession
	a.denyApproval = a.controller.DenyApproval
	a.moveResume = a.moveResumeSelection
	a.confirmResume = a.confirmResumeSelection
	a.cancelResume = func() { a.resumeOpen.Set(false) }
	return a
}

func (a *App) AttachApp(runtime *app.App) {
	a.quit = runtime.Stop
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

func (a *App) Profile() string {
	return a.controller.Profile()
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

func (a *App) CommandCompletion() components.CommandCompletion {
	if a.commands == nil || a.editor == nil || a.controller.Running() {
		return components.CommandCompletion{}
	}
	value := strings.TrimLeft(a.editor.Value().Value, " \t\r\n")
	if !strings.HasPrefix(value, "/") || strings.ContainsAny(value, " \t\r\n") {
		return components.CommandCompletion{}
	}
	prefix := strings.TrimPrefix(value, "/")
	matched := a.commands.Complete(prefix)
	if len(matched) == 0 {
		return components.CommandCompletion{}
	}
	if len(matched) > 8 {
		matched = matched[:8]
	}
	if prefix != a.completionPrefix {
		a.completionPrefix, a.completionSelected = prefix, 0
	}
	if a.completionSelected >= len(matched) {
		a.completionSelected = 0
	}
	candidates := make([]components.CommandCandidate, len(matched))
	for i, command := range matched {
		candidates[i] = components.CommandCandidate{Name: command.Name, Description: command.Description}
	}
	return components.CommandCompletion{Open: true, Candidates: candidates, Selected: a.completionSelected}
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
	if completion := a.CommandCompletion(); completion.Open {
		switch event.Key {
		case input.KeyArrowUp:
			a.moveCompletion(-1, len(completion.Candidates))
			return true
		case input.KeyArrowDown:
			a.moveCompletion(1, len(completion.Candidates))
			return true
		case input.KeyEnter, input.KeyTab:
			a.completeCommand(completion)
			return true
		}
	}
	if event.Key == input.KeyTab {
		a.toggleProfile()
		return true
	}
	if event.Key == input.KeyEnter && event.Modifiers == 0 {
		content := strings.TrimSpace(a.editor.Value().Value)
		if content == "" {
			return false
		}
		state := a.editor.Value()
		a.submit(content, &state)
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

func (a *App) moveCompletion(delta, count int) {
	if count == 0 {
		return
	}
	a.completionSelected = (a.completionSelected + delta + count) % count
	a.editor.Set(a.editor.Value())
}

func (a *App) completeCommand(completion components.CommandCompletion) {
	if completion.Selected < 0 || completion.Selected >= len(completion.Candidates) {
		return
	}
	a.setEditor("/" + completion.Candidates[completion.Selected].Name + " ")
	a.completionPrefix, a.completionSelected = "", 0
}

func (a *App) submit(content string, editor *builtin.TextareaState) {
	parsed, commandInput, err := domain.ParseCommand(content)
	if !commandInput {
		a.ensureSession()
		a.controller.Start(content, editor)
		return
	}
	if err != nil {
		a.controller.AddNotice("Invalid slash command. Type /help for available commands.")
		a.resetComposer(editor)
		return
	}
	if a.commands == nil {
		a.controller.AddNotice("Slash commands are unavailable.")
		a.resetComposer(editor)
		return
	}
	command := a.commands.Find(parsed.Name)
	if command == nil {
		a.controller.AddNotice("Unknown command: /" + parsed.Name + ". Type /help for available commands.")
		a.resetComposer(editor)
		return
	}
	if command.ArgumentHint != "" && parsed.Args == "" {
		a.controller.AddNotice("Usage: /" + command.Name + " " + command.ArgumentHint)
		a.resetComposer(editor)
		return
	}
	if command.Kind == domain.CommandPrompt {
		a.ensureSession()
		a.controller.StartSubmission(controller.Submission{Display: content, Prompt: command.RenderPrompt(parsed.Args)}, editor)
		return
	}
	a.executeLocal(command.Name, parsed.Args)
	a.resetComposer(editor)
}

func (a *App) ensureSession() {
	if session := a.controller.Session(); session != nil && session.ID() == "" {
		session.ConfigurePersistence(a.rootDir)
	}
}

func (a *App) resetComposer(editor *builtin.TextareaState) {
	*editor = builtin.TextareaState{PreferredColumn: -1, CursorOn: true}
	a.completionPrefix, a.completionSelected = "", 0
}

func (a *App) executeLocal(name, args string) {
	if a.controller.Running() {
		a.controller.AddNotice("Wait for the current operation to finish before running a local command.")
		return
	}
	switch name {
	case "help":
		if args != "" {
			command := a.commands.Find(args)
			if command == nil {
				a.controller.AddNotice("Unknown command: /" + args)
				return
			}
			a.controller.AddNotice(commandHelp(*command))
			return
		}
		var lines []string
		for _, command := range a.commands.List() {
			lines = append(lines, "/"+command.Name+" - "+command.Description)
		}
		a.controller.AddNotice("Available commands:\n\n" + strings.Join(lines, "\n") + "\n\nType /help <command> for details.")
	case "status":
		metrics := a.controller.ContextMetrics()
		a.controller.AddNotice(fmt.Sprintf("Codefolio status\n\nProfile: %s\nModel: %s\nContext: %d / %d tokens\nDirectory: %s", a.controller.Profile(), a.controller.ModelName(), metrics.UsedInputTokens(), metrics.UsableInputTokens(), a.workDir))
	case "mcp":
		a.controller.AddNotice(a.mcpSummary())
	case "new":
		if a.sessions == nil {
			a.controller.AddNotice("Session creation is unavailable.")
			return
		}
		value := a.sessions.New(a.systemPrompt)
		a.controller.ReplaceSession(value, nil)
		a.viewport.Set(builtin.ViewportState{FollowEnd: true})
	case "compact":
		a.compact()
	case "memory":
		a.memoryCommand(args)
	case "session":
		a.sessionCommand(args)
	case "resume":
		if strings.TrimSpace(args) == "" {
			a.openResume()
			return
		}
		a.resume(args)
	case "commands":
		if strings.TrimSpace(args) != "reload" {
			a.controller.AddNotice("Usage: /commands reload")
			return
		}
		loaded := commandinfra.Load(a.rootDir)
		diagnostics := a.commands.ReplaceDynamic(loaded.Commands)
		a.controller.AddNotice(fmt.Sprintf("Reloaded %d command(s); %d file diagnostic(s), %d registry conflict(s).", len(loaded.Commands), len(loaded.Diagnostics), len(diagnostics)))
	case "exit", "quit":
		if a.quit != nil {
			a.quit()
		}
	}
}

func (a *App) compact() {
	if a.contextManager == nil || a.toolRegistry == nil {
		a.controller.AddNotice("Context compaction is unavailable.")
		return
	}
	result, err := a.contextManager.Compact(context.Background(), a.controller.Session().ProviderMessages(), a.toolRegistry.GetAllSchemas(), a.cfg)
	if err != nil {
		a.controller.AddNotice("Compaction failed: " + err.Error())
		return
	}
	if !result.DidCompact {
		a.controller.AddNotice("No completed turns are eligible for compaction.")
		return
	}
	a.controller.Session().ReplaceProviderMessages(result.Messages)
	a.controller.SetContextMetrics(result.Metrics)
	a.controller.AddNotice(result.Detail)
}

func (a *App) memoryCommand(args string) {
	if a.memory == nil {
		a.controller.AddNotice("Memory management is unavailable.")
		return
	}
	switch strings.TrimSpace(args) {
	case "", "list":
		entries := a.memory.List(a.rootDir)
		if len(entries) == 0 {
			a.controller.AddNotice("No memories stored yet.")
			return
		}
		var lines []string
		for _, entry := range entries {
			description := entry.Description
			if description == "" {
				description = entry.Name
			}
			lines = append(lines, entry.ID+" - "+description)
		}
		a.controller.AddNotice("Memories:\n\n" + strings.Join(lines, "\n"))
	case "clear":
		removed, err := a.memory.Clear(a.rootDir)
		if err != nil {
			a.controller.AddNotice("Memory clear failed: " + err.Error())
			return
		}
		a.controller.AddNotice(fmt.Sprintf("Cleared %d memory file(s).", removed))
	default:
		a.controller.AddNotice("Usage: /memory [list|clear]")
	}
}

func (a *App) sessionCommand(args string) {
	if a.sessions == nil {
		a.controller.AddNotice("Session management is unavailable.")
		return
	}
	switch strings.TrimSpace(args) {
	case "", "info":
		value := a.controller.Session()
		a.controller.AddNotice(fmt.Sprintf("Session %s\n\nMessages: %d\nDuration: %s", value.ID(), value.MessageCount(), value.Duration().Round(time.Second)))
	case "list":
		items, err := a.sessions.List(context.Background(), a.rootDir)
		if err != nil {
			a.controller.AddNotice("List sessions failed: " + err.Error())
			return
		}
		if len(items) == 0 {
			a.controller.AddNotice("No previous sessions found.")
			return
		}
		var lines []string
		for _, item := range items {
			lines = append(lines, fmt.Sprintf("%s - %s (%d messages)", item.ID, item.Title, item.MessageCount))
		}
		a.controller.AddNotice("Sessions:\n\n" + strings.Join(lines, "\n"))
	case "new":
		a.executeLocal("new", "")
	default:
		a.controller.AddNotice("Usage: /session [info|list|new]")
	}
}

func (a *App) resume(id string) {
	if a.sessions == nil {
		a.controller.AddNotice("Session resume is unavailable.")
		return
	}
	var value domain.Session
	var info domain.SessionInfo
	var err error
	if strings.TrimSpace(id) == "" {
		value, info, err = a.sessions.Resume(context.Background(), a.rootDir)
	} else {
		value, info, err = a.sessions.Load(context.Background(), a.rootDir, strings.TrimSpace(id))
	}
	if err != nil {
		a.controller.AddNotice("Resume failed: " + err.Error())
		return
	}
	a.controller.ReplaceSession(value, controller.HydrateMessages(value.Messages(), a.controller.Profile()))
	a.viewport.Set(builtin.ViewportState{FollowEnd: true})
	a.editor.Set(builtin.TextareaState{PreferredColumn: -1, CursorOn: true})
	a.completionPrefix, a.completionSelected = "", 0
	a.controller.AddNotice("Resumed session " + info.ID + ".")
}

func (a *App) openResume() {
	if a.sessions == nil {
		a.controller.AddNotice("Session resume is unavailable.")
		return
	}
	items, err := a.sessions.List(context.Background(), a.rootDir)
	if err != nil {
		a.controller.AddNotice("List sessions failed: " + err.Error())
		return
	}
	if len(items) == 0 {
		a.controller.AddNotice("No previous sessions found.")
		return
	}
	a.resumeSessions, a.resumeSelected = items, 0
	a.resumeOpen.Set(true)
}

func (a *App) moveResumeSelection(delta int) {
	if len(a.resumeSessions) == 0 {
		return
	}
	a.resumeSelected = (a.resumeSelected + delta + len(a.resumeSessions)) % len(a.resumeSessions)
	a.resumeOpen.Set(true)
}

func (a *App) confirmResumeSelection() {
	if a.resumeSelected < 0 || a.resumeSelected >= len(a.resumeSessions) {
		return
	}
	id := a.resumeSessions[a.resumeSelected].ID
	a.resumeOpen.Set(false)
	a.resume(id)
}

func (a *App) mcpSummary() string {
	if a.mcpManager == nil {
		return "No MCP servers configured."
	}
	summary := a.mcpManager.Summary()
	if summary.Configured == 0 {
		return "No MCP servers configured."
	}
	return fmt.Sprintf("MCP: %d configured, %d ready, %d unavailable, %d tools.", summary.Configured, summary.Ready, summary.Unavailable, summary.Tools)
}

func commandHelp(command domain.Command) string {
	var output strings.Builder
	fmt.Fprintf(&output, "/%s - %s", command.Name, command.Description)
	if len(command.Aliases) > 0 {
		fmt.Fprintf(&output, "\nAliases: /%s", strings.Join(command.Aliases, ", /"))
	}
	if command.ArgumentHint != "" {
		fmt.Fprintf(&output, "\nUsage: /%s %s", command.Name, command.ArgumentHint)
	}
	return output.String()
}

func (a *App) toggleProfile() {
	if a.controller.Mode() == svc.ModePlan {
		a.controller.SetMode(svc.ModeDefault)
		return
	}
	a.controller.SetMode(svc.ModePlan)
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
