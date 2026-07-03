package tux

// Key represents a keyboard key.
type Key uint16

const (
	KeyNone Key = iota
	KeyRune // Printable character (see KeyboardEvent.Rune)

	// Special keys
	KeyEsc
	KeyTab
	KeyEnter
	KeyBackspace
	KeyDelete

	// Arrow keys
	KeyUp
	KeyDown
	KeyLeft
	KeyRight

	// Future: Home, End, PageUp, PageDown, F1-F12...
)

// Modifier represents keyboard modifier flags.
type Modifier uint8

const (
	ModNone  Modifier = 0
	ModCtrl  Modifier = 1 << iota
	ModAlt
	ModShift
)

// KeyboardEvent represents a keyboard event.
type KeyboardEvent struct {
	Key  Key      // Key pressed (KeyRune for characters, or specific key)
	Rune rune     // Character (valid when Key == KeyRune)
	Mod  Modifier // Modifier flags (bitwise combination)
}

// KeyboardHandler handles keyboard events.
type KeyboardHandler interface {
	OnKeyboard(e KeyboardEvent) error
}

// ClickEvent represents a mouse click event (reserved for future use).
type ClickEvent struct {
	Row       int
	Column    int
	Button    int      // 0=left, 1=middle, 2=right
	Modifiers Modifier // Modifiers held during click
}

// ClickHandler handles click events (reserved for future use).
type ClickHandler interface {
	OnClick(e ClickEvent) error
}
