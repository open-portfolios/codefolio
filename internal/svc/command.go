package svc

import (
	"sort"
	"strings"
	"sync"

	"github.com/open-portfolios/codefolio/internal/domain"
)

type CommandRegistry struct {
	mu       sync.RWMutex
	builtins []domain.Command
	dynamic  []domain.Command
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{builtins: builtinCommands()}
}

func (r *CommandRegistry) ReplaceDynamic(commands []domain.Command) []domain.CommandDiagnostic {
	r.mu.Lock()
	defer r.mu.Unlock()
	accepted, diagnostics := resolveCommands(append(append([]domain.Command(nil), r.builtins...), commands...))
	if len(accepted) < len(r.builtins) {
		return diagnostics
	}
	r.dynamic = accepted[len(r.builtins):]
	return diagnostics
}

func (r *CommandRegistry) Find(name string) *domain.Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	name = strings.ToLower(name)
	for _, command := range append(append([]domain.Command(nil), r.builtins...), r.dynamic...) {
		if command.Name == name {
			copy := command
			return &copy
		}
		for _, alias := range command.Aliases {
			if strings.ToLower(alias) == name {
				copy := command
				return &copy
			}
		}
	}
	return nil
}

func (r *CommandRegistry) List() []domain.Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	commands := append(append([]domain.Command(nil), r.builtins...), r.dynamic...)
	sort.Slice(commands, func(i, j int) bool { return commands[i].Name < commands[j].Name })
	return commands
}

func (r *CommandRegistry) Complete(prefix string) []domain.Command {
	prefix = strings.ToLower(prefix)
	var result []domain.Command
	for _, command := range r.List() {
		if strings.HasPrefix(command.Name, prefix) {
			result = append(result, command)
		}
	}
	return result
}

func resolveCommands(commands []domain.Command) ([]domain.Command, []domain.CommandDiagnostic) {
	accepted := make([]domain.Command, 0, len(commands))
	used := map[string]struct{}{}
	for _, command := range commands {
		command.Name = strings.ToLower(command.Name)
		if !domain.ValidCommandName(command.Name) {
			continue
		}
		conflict := false
		keys := append([]string{command.Name}, command.Aliases...)
		local := map[string]struct{}{}
		for index, key := range keys {
			key = strings.ToLower(key)
			if !domain.ValidCommandName(key) || (index > 0 && key == command.Name) {
				conflict = true
			}
			if _, exists := used[key]; exists {
				conflict = true
			}
			if _, exists := local[key]; exists {
				conflict = true
			}
			local[key] = struct{}{}
		}
		if conflict {
			continue
		}
		for key := range local {
			used[key] = struct{}{}
		}
		accepted = append(accepted, command)
	}
	return accepted, nil
}

func builtinCommands() []domain.Command {
	local := func(name, description string, aliases ...string) domain.Command {
		return domain.Command{Name: name, Description: description, Aliases: aliases, Kind: domain.CommandLocal, Source: domain.CommandBuiltin}
	}
	return []domain.Command{
		local("help", "Show available commands", "h"),
		local("status", "Show current status", "s"),
		local("mcp", "Show MCP server status"),
		local("new", "Start a new session"),
		local("compact", "Compact provider conversation context", "c"),
		local("memory", "List or clear persistent memory"),
		local("session", "Show or list sessions"),
		local("resume", "Resume a previous session", "r"),
		local("commands", "Reload Markdown prompt commands"),
		local("exit", "Exit Codefolio"),
		local("quit", "Exit Codefolio"),
		{Name: "review", Description: "Review current code changes", ArgumentHint: "[focus]", Kind: domain.CommandPrompt, Source: domain.CommandBuiltin, RenderPrompt: func(args string) string {
			prompt := "Review the current git diff for logic errors, security issues, performance problems, and missing tests. Report actionable findings first."
			if strings.TrimSpace(args) != "" {
				prompt += "\n\nAdditional focus: " + args
			}
			return prompt
		}},
	}
}
