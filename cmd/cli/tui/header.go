package tui

import (
	"os"
	"path/filepath"

	"charm.land/lipgloss/v2"
)

func renderHeader(model, workDir string, width int) string {
	home, _ := os.UserHomeDir()
	displayDir := workDir
	if home != "" {
		if rel, err := filepath.Rel(home, workDir); err == nil && len(rel) < len(displayDir) {
			displayDir = "~/" + filepath.ToSlash(rel)
		}
	}

	brand := lipgloss.NewStyle().
		Background(darkBg).
		Foreground(accent).
		Bold(true).
		Render("codefolio")

	info := lipgloss.NewStyle().
		Background(darkBg).
		Foreground(lipgloss.Color("#9CA3AF")).
		Render(model)

	pathPart := lipgloss.NewStyle().
		Background(darkBg).
		Foreground(mutedFg).
		Render(displayDir)

	sep := lipgloss.NewStyle().
		Background(darkBg).
		Foreground(lipgloss.Color("#4B5563")).
		Render(" · ")

	content := lipgloss.JoinHorizontal(lipgloss.Left,
		" ", brand, sep, info, sep, pathPart,
	)

	return headerBaseStyle.Width(width).Render(content)
}

var headerBaseStyle = lipgloss.NewStyle().
	Background(darkBg).
	Foreground(lipgloss.Color("#D1D5DB")).
	Padding(0, 1)
