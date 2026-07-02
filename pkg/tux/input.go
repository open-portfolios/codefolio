package tux

type Key int

const (
	KeyRune Key = iota
	KeyEsc
	KeyBackspace
	KeyDelete
	KeyEnter
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
)

type InputEvent struct {
	Key  Key
	Rune rune
}
