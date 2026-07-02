package tux

type Key int

const (
	KeyByte Key = iota
	KeyEsc
	KeyBackspace
	KeyEnter
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
)

type InputEvent struct {
	Key  Key
	Byte byte
}
