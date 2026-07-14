package bash

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/tools/shared"
)

const maxTimeout = 600

type Tool struct{}

func (t *Tool) Name() string                  { return "Bash" }
func (t *Tool) Description() string           { return shared.BashDescription }
func (t *Tool) Category() domain.ToolCategory { return domain.CategoryCommand }

func (t *Tool) Schema() map[string]any {
	return map[string]any{
		"name":        t.Name(),
		"description": t.Description(),
		"input_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{"type": "string", "description": "Shell command to execute"},
				"timeout": map[string]any{"type": "integer", "description": "Timeout in seconds (max 600)", "default": 120},
			},
			"required": []string{"command"},
		},
	}
}

func (t *Tool) Execute(ctx context.Context, args map[string]any) domain.ToolResult {
	command, _ := args["command"].(string)
	if command == "" {
		return domain.ToolResult{Output: "Error: command is required", IsError: true}
	}

	timeout := min(shared.IntArg(args, "timeout", 120), maxTimeout)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return domain.ToolResult{Output: fmt.Sprintf("Error: command timed out after %ds", timeout), IsError: true}
	}

	exitCode := 0
	isError := false
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			isError = exitCode != 0
		} else if ctx.Err() == nil {
			return domain.ToolResult{Output: fmt.Sprintf("Error executing command: %s", err), IsError: true}
		}
	}

	var sb bytes.Buffer
	fmt.Fprintf(&sb, "$ %s\n", command)
	if stdout.Len() > 0 {
		sb.Write(stdout.Bytes())
		if stdout.Bytes()[stdout.Len()-1] != '\n' {
			sb.WriteByte('\n')
		}
	}
	if stderr.Len() > 0 {
		fmt.Fprintf(&sb, "STDERR: %s", stderr.String())
		if stderr.Bytes()[stderr.Len()-1] != '\n' {
			sb.WriteByte('\n')
		}
	}
	if exitCode != 0 {
		fmt.Fprintf(&sb, "(exit code %d)", exitCode)
	}

	return domain.ToolResult{Output: sb.String(), IsError: isError}
}
