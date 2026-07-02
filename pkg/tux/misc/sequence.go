package misc

import "fmt"

// ColorSequence returns an ANSI SGR escape sequence that sets the foreground
// and background colors. ColorDefault resets the respective attribute.
func ColorSequence(foreground, background Color) string {
	return fmt.Sprintf("\x1b[%d;%dm", AnsiForeground(foreground), AnsiBackground(background))
}

// AnsiForeground maps a Color value to its ANSI SGR foreground code.
func AnsiForeground(color Color) int {
	switch color {
	case ColorBlack:
		return 30
	case ColorRed:
		return 31
	case ColorGreen:
		return 32
	case ColorYellow:
		return 33
	case ColorBlue:
		return 34
	case ColorMagenta:
		return 35
	case ColorCyan:
		return 36
	case ColorWhite:
		return 37
	default:
		return 39
	}
}

// AnsiBackground maps a Color value to its ANSI SGR background code.
func AnsiBackground(color Color) int {
	switch color {
	case ColorBlack:
		return 40
	case ColorRed:
		return 41
	case ColorGreen:
		return 42
	case ColorYellow:
		return 43
	case ColorBlue:
		return 44
	case ColorMagenta:
		return 45
	case ColorCyan:
		return 46
	case ColorWhite:
		return 47
	default:
		return 49
	}
}
