package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

func renderSidebar(session *Session, width int) string {
	sections := []string{
		SidebarTitleStyle.Render("Session"),
		"",
		sidebarRow("Messages", fmt.Sprintf("%d", session.MessageCount())),
		sidebarRow("Context", fmt.Sprintf("%.0f%%", session.ContextUsage())),
		sidebarRow("LSP", "—"),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return SidebarStyle.Width(width).Render(content)
}

func sidebarRow(label, value string) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		SidebarLabelStyle.Render(label+":"),
		"  ",
		SidebarValueStyle.Render(value),
	)
}
