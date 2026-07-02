package misc

// Color represents an ANSI terminal color. The zero value is ColorDefault,
// which resets the terminal color to its default.
type Color uint8

const (
	ColorDefault Color = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)
