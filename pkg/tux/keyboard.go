package tux

import (
	"time"
	"unicode/utf8"
)

// keyboardListener decodes raw stdin bytes into KeyboardEvents. It handles
// multi-byte UTF-8 sequences, ANSI escape sequences (arrow keys, Delete),
// Windows extended key codes (0x00 / 0xE0 prefix), and standalone ESC presses.
type keyboardListener struct {
	escPending  bool
	escBuffer   []byte
	escDeadline time.Time
	extPending  bool
	utf8Buffer  []byte
}

// handleByte processes one raw byte from stdin and returns any KeyboardEvents it
// produces. Most bytes produce at most one event; an ESC byte produces none
// until either the sequence completes or the deadline expires (see flush).
func (k *keyboardListener) handleByte(b byte) []KeyboardEvent {
	if k.extPending {
		k.extPending = false
		switch b {
		case 'H':
			return []KeyboardEvent{{Key: KeyUp}}
		case 'P':
			return []KeyboardEvent{{Key: KeyDown}}
		case 'K':
			return []KeyboardEvent{{Key: KeyLeft}}
		case 'M':
			return []KeyboardEvent{{Key: KeyRight}}
		}
		return nil
	}

	if k.escPending {
		k.escBuffer = append(k.escBuffer, b)

		// Waiting for '[' as the first byte after ESC.
		if len(k.escBuffer) == 1 {
			if b != '[' {
				// Not a CSI sequence — discard and emit standalone ESC.
				k.escPending = false
				k.escBuffer = nil
				return []KeyboardEvent{{Key: KeyEsc}}
			}
			return nil
		}

		// CSI sequence: ESC [ <byte>
		if len(k.escBuffer) == 2 {
			if k.escBuffer[1] >= '0' && k.escBuffer[1] <= '9' {
				// Extended sequence like ESC [ 3 ~ — wait for one more byte.
				return nil
			}
			var event *KeyboardEvent
			switch k.escBuffer[1] {
			case 'A':
				event = &KeyboardEvent{Key: KeyUp}
			case 'B':
				event = &KeyboardEvent{Key: KeyDown}
			case 'C':
				event = &KeyboardEvent{Key: KeyRight}
			case 'D':
				event = &KeyboardEvent{Key: KeyLeft}
			}
			k.escPending = false
			k.escBuffer = nil
			if event != nil {
				return []KeyboardEvent{*event}
			}
			return nil
		}

		// ESC [ <digit> <byte>
		if len(k.escBuffer) == 3 {
			var event *KeyboardEvent
			if k.escBuffer[1] == '3' && k.escBuffer[2] == '~' {
				event = &KeyboardEvent{Key: KeyDelete}
			}
			k.escPending = false
			k.escBuffer = nil
			if event != nil {
				return []KeyboardEvent{*event}
			}
			return nil
		}

		// Sequence longer than expected — discard.
		k.escPending = false
		k.escBuffer = nil
		return nil
	}

	switch b {
	case 0, 224:
		k.extPending = true
		return nil
	case 27:
		k.escPending = true
		k.escBuffer = nil
		k.escDeadline = time.Now().Add(30 * time.Millisecond)
		return nil
	case 8, 127:
		return []KeyboardEvent{{Key: KeyBackspace}}
	case '\r', '\n':
		return []KeyboardEvent{{Key: KeyEnter}}
	case '\t':
		return []KeyboardEvent{{Key: KeyTab}}
	case 1, 2, 3, 4, 5, 6, 7, 11, 12, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26:
		// Ctrl+A through Ctrl+Z (excluding tab, enter, backspace)
		// Map control characters to Ctrl + lowercase letter
		return []KeyboardEvent{{Key: KeyRune, Rune: rune('a' + b - 1), Mod: ModCtrl}}
	default:
		return k.handleRuneByte(b)
	}
}

// handleRuneByte assembles multi-byte UTF-8 sequences and emits a KeyRune
// event once a complete rune is decoded.
func (k *keyboardListener) handleRuneByte(b byte) []KeyboardEvent {
	if b < utf8.RuneSelf && len(k.utf8Buffer) == 0 {
		return []KeyboardEvent{{Key: KeyRune, Rune: rune(b)}}
	}

	k.utf8Buffer = append(k.utf8Buffer, b)
	if !utf8.FullRune(k.utf8Buffer) {
		return nil
	}
	r, size := utf8.DecodeRune(k.utf8Buffer)
	k.utf8Buffer = nil
	if r != utf8.RuneError || size > 1 {
		return []KeyboardEvent{{Key: KeyRune, Rune: r}}
	}
	return nil
}

// flush checks whether a pending ESC sequence has timed out. If the deadline
// has passed, the pending state is cleared and a standalone KeyEsc event is
// returned. Call this once per frame to bound ESC latency.
func (k *keyboardListener) flush() []KeyboardEvent {
	if !k.escPending || time.Now().Before(k.escDeadline) {
		return nil
	}
	k.escPending = false
	k.escBuffer = nil
	return []KeyboardEvent{{Key: KeyEsc}}
}
