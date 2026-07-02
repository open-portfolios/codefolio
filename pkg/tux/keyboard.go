package tux

import (
	"time"
	"unicode/utf8"
)

// keyboardListener decodes raw stdin bytes into InputEvents. It handles
// multi-byte UTF-8 sequences, ANSI escape sequences (arrow keys, Delete),
// Windows extended key codes (0x00 / 0xE0 prefix), and standalone ESC presses.
type keyboardListener struct {
	escPending  bool
	escBuffer   []byte
	escDeadline time.Time
	extPending  bool
	utf8Buffer  []byte
}

// handleByte processes one raw byte from stdin and returns any InputEvents it
// produces. Most bytes produce at most one event; an ESC byte produces none
// until either the sequence completes or the deadline expires (see flush).
func (k *keyboardListener) handleByte(b byte) []InputEvent {
	if k.extPending {
		k.extPending = false
		switch b {
		case 'H':
			return []InputEvent{{Key: KeyUp}}
		case 'P':
			return []InputEvent{{Key: KeyDown}}
		case 'K':
			return []InputEvent{{Key: KeyLeft}}
		case 'M':
			return []InputEvent{{Key: KeyRight}}
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
				return []InputEvent{{Key: KeyEsc}}
			}
			return nil
		}

		// CSI sequence: ESC [ <byte>
		if len(k.escBuffer) == 2 {
			if k.escBuffer[1] >= '0' && k.escBuffer[1] <= '9' {
				// Extended sequence like ESC [ 3 ~ — wait for one more byte.
				return nil
			}
			var event *InputEvent
			switch k.escBuffer[1] {
			case 'A':
				event = &InputEvent{Key: KeyUp}
			case 'B':
				event = &InputEvent{Key: KeyDown}
			case 'C':
				event = &InputEvent{Key: KeyRight}
			case 'D':
				event = &InputEvent{Key: KeyLeft}
			}
			k.escPending = false
			k.escBuffer = nil
			if event != nil {
				return []InputEvent{*event}
			}
			return nil
		}

		// ESC [ <digit> <byte>
		if len(k.escBuffer) == 3 {
			var event *InputEvent
			if k.escBuffer[1] == '3' && k.escBuffer[2] == '~' {
				event = &InputEvent{Key: KeyDelete}
			}
			k.escPending = false
			k.escBuffer = nil
			if event != nil {
				return []InputEvent{*event}
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
		return []InputEvent{{Key: KeyBackspace, Rune: rune(b)}}
	case '\r', '\n':
		return []InputEvent{{Key: KeyEnter, Rune: rune(b)}}
	default:
		return k.handleRuneByte(b)
	}
}

// handleRuneByte assembles multi-byte UTF-8 sequences and emits a KeyRune
// event once a complete rune is decoded.
func (k *keyboardListener) handleRuneByte(b byte) []InputEvent {
	if b < utf8.RuneSelf && len(k.utf8Buffer) == 0 {
		return []InputEvent{{Key: KeyRune, Rune: rune(b)}}
	}

	k.utf8Buffer = append(k.utf8Buffer, b)
	if !utf8.FullRune(k.utf8Buffer) {
		return nil
	}
	r, size := utf8.DecodeRune(k.utf8Buffer)
	k.utf8Buffer = nil
	if r != utf8.RuneError || size > 1 {
		return []InputEvent{{Key: KeyRune, Rune: r}}
	}
	return nil
}

// flush checks whether a pending ESC sequence has timed out. If the deadline
// has passed, the pending state is cleared and a standalone KeyEsc event is
// returned. Call this once per frame to bound ESC latency.
func (k *keyboardListener) flush() []InputEvent {
	if !k.escPending || time.Now().Before(k.escDeadline) {
		return nil
	}
	k.escPending = false
	k.escBuffer = nil
	return []InputEvent{{Key: KeyEsc}}
}
