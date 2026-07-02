package tux

import "github.com/rivo/uniseg"

func RuneWidth(r rune) int {
	if r == 0 {
		return 0
	}
	width := uniseg.StringWidth(string(r))
	if width < 1 {
		return 1
	}
	if width > 2 {
		return 2
	}
	return width
}

func StringWidth(s string) int {
	return uniseg.StringWidth(s)
}

func NewCell(ch rune, foreground, background Color) Cell {
	return Cell{
		Ch:         ch,
		Width:      uint8(RuneWidth(ch)),
		Foreground: foreground,
		Background: background,
	}
}

func ContinuationCell(foreground, background Color) Cell {
	return Cell{
		Width:      0,
		Foreground: foreground,
		Background: background,
	}
}
