package tux

import (
	"bufio"
	"fmt"
	"os"
)

var (
	_ RenderContext = (*Renderer)(nil)
)

type Renderer struct {
	buf          [][]Cell
	cursorQueued bool
	cursorRow    int
	cursorColumn int
}

func NewRenderer(row, column int) *Renderer {
	buf := make([][]Cell, row)
	for i := range buf {
		buf[i] = make([]Cell, column)
	}
	return &Renderer{buf: buf}
}

func (r *Renderer) Clear() {
	for row := range r.buf {
		for column := range r.buf[row] {
			r.buf[row][column] = Cell{}
		}
	}
	r.cursorQueued = false
	r.cursorRow = 0
	r.cursorColumn = 0
}

func (r *Renderer) Render(ctx BuildContext, root Component) error {
	if err := root.Render(ctx, r); err != nil {
		return err
	}
	return r.Submit()
}

func (r *Renderer) Paint(row, column int, cell Cell) { r.buf[row][column] = cell }

func (r *Renderer) QueueCursorMove(row, column int) {
	r.cursorQueued = true
	r.cursorRow = row
	r.cursorColumn = column
}

func (r *Renderer) Submit() error {
	w := bufio.NewWriter(os.Stdout)
	if _, err := w.WriteString("\x1b[H"); err != nil {
		return err
	}
	foreground := ColorDefault
	background := ColorDefault
	for row, cells := range r.buf {
		if row > 0 {
			if foreground != ColorDefault || background != ColorDefault {
				if _, err := w.WriteString("\x1b[0m"); err != nil {
					return err
				}
				foreground = ColorDefault
				background = ColorDefault
			}
			if err := w.WriteByte('\n'); err != nil {
				return err
			}
		}

		for _, cell := range cells {
			if cell.Width == 0 {
				continue
			}

			if cell.Foreground != foreground || cell.Background != background {
				if _, err := w.WriteString(colorSequence(cell.Foreground, cell.Background)); err != nil {
					return err
				}
				foreground = cell.Foreground
				background = cell.Background
			}

			ch := cell.Ch
			if ch == 0 {
				ch = ' '
			}
			if _, err := w.WriteRune(ch); err != nil {
				return err
			}
		}
	}
	if foreground != ColorDefault || background != ColorDefault {
		if _, err := w.WriteString("\x1b[0m"); err != nil {
			return err
		}
	}

	if r.cursorQueued {
		if _, err := fmt.Fprintf(w, "\x1b[%d;%dH", r.cursorRow+1, r.cursorColumn+1); err != nil {
			return err
		}
	}

	return w.Flush()
}

func colorSequence(foreground, background Color) string {
	return fmt.Sprintf("\x1b[%d;%dm", ansiForeground(foreground), ansiBackground(background))
}

func ansiForeground(color Color) int {
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

func ansiBackground(color Color) int {
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
