package domain

import "strings"

type CommandKind uint8

const (
	CommandLocal CommandKind = iota
	CommandPrompt
)

type CommandSource uint8

const (
	CommandBuiltin CommandSource = iota
	CommandUser
	CommandProject
)

type Command struct {
	Name         string
	Description  string
	Aliases      []string
	ArgumentHint string
	Kind         CommandKind
	Source       CommandSource
	Path         string
	RenderPrompt func(string) string
}

type ParsedCommand struct {
	Name string
	Args string
}

type CommandDiagnostic struct {
	Path    string
	Message string
}

// ParseCommand recognizes a single slash-prefixed command. Arguments remain
// raw because prompt commands intentionally receive the user's original text.
func ParseCommand(input string) (ParsedCommand, bool, error) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return ParsedCommand{}, false, nil
	}
	if strings.HasPrefix(input, "//") {
		return ParsedCommand{}, true, ErrInvalidCommand
	}
	input = strings.TrimPrefix(input, "/")
	index := strings.IndexFunc(input, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' })
	name, args, found := input, "", false
	if index >= 0 {
		name, args, found = input[:index], input[index:], true
	}
	if name == "" || !ValidCommandName(name) {
		return ParsedCommand{}, true, ErrInvalidCommand
	}
	if found {
		args = strings.TrimSpace(args)
	}
	return ParsedCommand{Name: strings.ToLower(name), Args: args}, true, nil
}

func ValidCommandName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == ':' {
			continue
		}
		return false
	}
	return true
}
