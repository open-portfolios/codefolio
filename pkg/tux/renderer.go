package tux

import (
	"bufio"
	"fmt"
	"os"

	"github.com/open-portfolios/codefolio/pkg/tux/misc"
)

var (
	_ RenderContext = (*Renderer)(nil)
)

// Renderer is the production renderer that writes to stdout using ANSI escape
// sequences. It keeps two cell grids: cur is written during the current frame
// and prev holds the last frame sent to the terminal. Submit diffs the two and
// only emits escape sequences for cells that changed, minimising terminal I/O.
type Renderer struct {
	prev         [][]Cell // last frame sent to terminal
	cur          [][]Cell // current frame being built
	cursorQueued bool
	cursorRow    int
	cursorColumn int
}

func NewRenderer(rows, columns int) *Renderer {
	return &Renderer{
		prev: makeCellGrid(rows, columns),
		cur:  makeCellGrid(rows, columns),
	}
}

func makeCellGrid(rows, columns int) [][]Cell {
	grid := make([][]Cell, rows)
	for i := range grid {
		grid[i] = make([]Cell, columns)
	}
	return grid
}

// Clear resets the current frame to zero cells. prev is untouched so the next
// Submit can diff correctly against the last submitted frame.
func (r *Renderer) Clear() {
	for row := range r.cur {
		for col := range r.cur[row] {
			r.cur[row][col] = Cell{}
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

func (r *Renderer) Paint(row, column int, cell Cell) { r.cur[row][column] = cell }

func (r *Renderer) QueueCursorMove(row, column int) {
	r.cursorQueued = true
	r.cursorRow = row
	r.cursorColumn = column
}

// Submit diffs cur against prev and writes only changed cells to stdout.
// Each changed cell gets an absolute cursor-positioning sequence so cells can
// be emitted in any order without redrawing unaffected content. After flushing,
// cur is copied into prev for the next frame.
func (r *Renderer) Submit() error {
	rows := len(r.cur)
	if rows == 0 {
		return nil
	}
	cols := len(r.cur[0])

	w := bufio.NewWriter(os.Stdout)

	// termRow/termCol track where the terminal cursor currently is so
	// consecutive changed cells in the same row avoid redundant moves.
	// -1 means "position unknown".
	termRow := -1
	termCol := -1
	fg := misc.ColorDefault
	bg := misc.ColorDefault

	for row := range rows {
		for col := range cols {
			newCell := r.cur[row][col]
			oldCell := r.prev[row][col]

			if newCell == oldCell {
				// Unchanged — skip, but we lose track of cursor position
				// because the terminal cursor did not advance here.
				if termRow == row && termCol == col {
					termRow = -1
					termCol = -1
				}
				continue
			}

			// Wide-char continuation slots (Width == 0) are implicitly
			// updated when the leading cell is written; skip them here.
			if newCell.Width == 0 {
				continue
			}

			// Move cursor if not already in position.
			if termRow != row || termCol != col {
				if _, err := fmt.Fprintf(w, "\x1b[%d;%dH", row+1, col+1); err != nil {
					return err
				}
				termRow = row
				termCol = col
			}

			// Emit color change only when necessary.
			if newCell.Foreground != fg || newCell.Background != bg {
				if _, err := w.WriteString(misc.ColorSequence(newCell.Foreground, newCell.Background)); err != nil {
					return err
				}
				fg = newCell.Foreground
				bg = newCell.Background
			}

			ch := newCell.Ch
			if ch == 0 {
				ch = ' '
			}
			if _, err := w.WriteRune(ch); err != nil {
				return err
			}

			// Advance tracked cursor position.
			// Do NOT assume the terminal auto-wraps when we reach the end of
			// the buffer row: the physical terminal may be wider than the
			// buffer, so the cursor would land at (row, cols) rather than
			// (row+1, 0). Reset to unknown to force an explicit move for the
			// next cell in a different row.
			advance := max(int(newCell.Width), 1)
			termCol += advance
			if termCol >= cols {
				termRow = -1
				termCol = -1
			}
		}
	}

	// Reset colors if we left the terminal in a non-default state.
	if fg != misc.ColorDefault || bg != misc.ColorDefault {
		if _, err := w.WriteString("\x1b[0m"); err != nil {
			return err
		}
	}

	// Apply queued cursor move (e.g. Input component placing the edit cursor).
	if r.cursorQueued {
		if _, err := fmt.Fprintf(w, "\x1b[%d;%dH", r.cursorRow+1, r.cursorColumn+1); err != nil {
			return err
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}

	// Copy cur → prev for the next frame's diff.
	for row := range r.cur {
		copy(r.prev[row], r.cur[row])
	}

	return nil
}
