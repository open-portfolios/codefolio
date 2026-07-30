package components

import "github.com/cylixlee/tux/style"

// Theme keeps the session chrome and transcript on a shared visual scale.
var Theme = struct {
	Background        style.Color
	BackgroundPanel   style.Color
	BackgroundElement style.Color
	Text              style.Color
	TextMuted         style.Color
	Border            style.Color
	Primary           style.Color
	Secondary         style.Color
	Accent            style.Color
	Success           style.Color
	Warning           style.Color
	Error             style.Color
}{
	Background:        style.HexColor("#0A0A0A"),
	BackgroundPanel:   style.HexColor("#141414"),
	BackgroundElement: style.HexColor("#1E1E1E"),
	Text:              style.HexColor("#EEEEEE"),
	TextMuted:         style.HexColor("#808080"),
	Border:            style.HexColor("#484848"),
	Primary:           style.HexColor("#FAB283"),
	Secondary:         style.HexColor("#5C9CF5"),
	Accent:            style.HexColor("#9D7CD8"),
	Success:           style.HexColor("#7FD88F"),
	Warning:           style.HexColor("#F5A742"),
	Error:             style.HexColor("#E06C75"),
}

func ProfileColor(profile string) style.Color {
	if profile == "plan" {
		return Theme.Primary
	}
	return Theme.Secondary
}
