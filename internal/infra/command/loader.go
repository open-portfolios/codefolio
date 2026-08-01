package command

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/open-portfolios/codefolio/internal/domain"
)

const maxCommandBytes = 256 * 1024

type Meta struct {
	Description  string   `yaml:"description"`
	ArgumentHint string   `yaml:"argument-hint"`
	Aliases      []string `yaml:"aliases"`
}

type LoadResult struct {
	Commands    []domain.Command
	Diagnostics []domain.CommandDiagnostic
}

func Load(workDir string) LoadResult {
	result := LoadDir(commandDir(homeDir()), domain.CommandUser)
	project := LoadDir(filepath.Join(workDir, ".codefolio", "commands"), domain.CommandProject)
	result.Diagnostics = append(result.Diagnostics, project.Diagnostics...)
	byName := make(map[string]int, len(result.Commands))
	for i, value := range result.Commands {
		byName[value.Name] = i
	}
	for _, value := range project.Commands {
		if index, ok := byName[value.Name]; ok {
			result.Commands[index] = value
			continue
		}
		byName[value.Name] = len(result.Commands)
		result.Commands = append(result.Commands, value)
	}
	return result
}

func LoadDir(root string, source domain.CommandSource) LoadResult {
	var result LoadResult
	if root == "" {
		return result
	}
	entries := make([]string, 0)
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			result.Diagnostics = append(result.Diagnostics, domain.CommandDiagnostic{Path: path, Message: err.Error()})
			return nil
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			entries = append(entries, path)
		}
		return nil
	})
	sort.Strings(entries)
	for _, path := range entries {
		value, err := parse(root, path, source)
		if err != nil {
			result.Diagnostics = append(result.Diagnostics, domain.CommandDiagnostic{Path: path, Message: err.Error()})
			continue
		}
		result.Commands = append(result.Commands, value)
	}
	return result
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".codefolio", "commands")
}

func commandDir(path string) string { return path }

func parse(root, path string, source domain.CommandSource) (domain.Command, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Command{}, err
	}
	if len(data) > maxCommandBytes {
		return domain.Command{}, fmt.Errorf("command exceeds %d KiB", maxCommandBytes/1024)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return domain.Command{}, fmt.Errorf("command is outside its root")
	}
	name := strings.TrimSuffix(rel, ".md")
	name = strings.ReplaceAll(name, string(filepath.Separator), ":")
	name = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	if !domain.ValidCommandName(name) {
		return domain.Command{}, fmt.Errorf("invalid command name %q", name)
	}
	meta, body, err := frontmatter(string(data))
	if err != nil {
		return domain.Command{}, err
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return domain.Command{}, fmt.Errorf("command body is empty")
	}
	if meta.Description == "" {
		meta.Description = description(body)
	}
	for _, alias := range meta.Aliases {
		if !domain.ValidCommandName(alias) {
			return domain.Command{}, fmt.Errorf("invalid alias %q", alias)
		}
	}
	return domain.Command{Name: name, Description: meta.Description, Aliases: meta.Aliases, ArgumentHint: meta.ArgumentHint, Kind: domain.CommandPrompt, Source: source, Path: path, RenderPrompt: render(body)}, nil
}

func frontmatter(content string) (Meta, string, error) {
	if !strings.HasPrefix(content, "---\n") {
		return Meta{}, content, nil
	}
	end := strings.Index(content[4:], "\n---\n")
	if end < 0 {
		return Meta{}, "", fmt.Errorf("frontmatter is not closed")
	}
	var meta Meta
	if err := yaml.Unmarshal([]byte(content[4:4+end]), &meta); err != nil {
		return Meta{}, "", fmt.Errorf("parse frontmatter: %w", err)
	}
	return meta, content[4+end+5:], nil
}

func description(body string) string {
	for line := range strings.SplitSeq(body, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return line
		}
	}
	return ""
}

func render(body string) func(string) string {
	return func(args string) string {
		if strings.Contains(body, "$ARGUMENTS") {
			return strings.ReplaceAll(body, "$ARGUMENTS", args)
		}
		if strings.TrimSpace(args) == "" {
			return body
		}
		return body + "\n\n## User Request\n\n" + args
	}
}
