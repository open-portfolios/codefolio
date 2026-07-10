package tui

import "charm.land/lipgloss/v2"

var (
	accent    = lipgloss.Color("#8B5CF6")
	mutedFg   = lipgloss.Color("#6B7280")
	borderFg  = lipgloss.Color("#374151")
	darkBg    = lipgloss.Color("#1F2937")
	userBarFg = lipgloss.Color("#818CF8")

	SidebarStyle = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(borderFg).
			Padding(0, 1).
			Width(30)

	SidebarTitleStyle = lipgloss.NewStyle().
				Foreground(accent).
				Bold(true)

	SidebarLabelStyle = lipgloss.NewStyle().
				Foreground(mutedFg)

	SidebarValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D1D5DB"))

	userMsgBar = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(userBarFg)

	thinkingStyle = lipgloss.NewStyle().
			Foreground(mutedFg).
			Italic(true)

	thinkingHeaderStyle = lipgloss.NewStyle().
				Foreground(accent)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true).
			PaddingLeft(2)
)
