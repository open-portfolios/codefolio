package builtin

import (
	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/misc"
)

var (
	_ tux.Component = (*box)(nil)
)

type box struct {
	tux.Atomic

	row        int
	column     int
	width      int
	height     int
	content    string
	foreground misc.Color
	background misc.Color
}

type BoxProps struct {
	Row        int
	Column     int
	Width      int
	Height     int
	Content    string
	Foreground misc.Color
	Background misc.Color
}

func Box(props BoxProps, children ...tux.Component) tux.Component {
	return &box{
		row:        props.Row,
		column:     props.Column,
		width:      props.Width,
		height:     props.Height,
		content:    props.Content,
		foreground: props.Foreground,
		background: props.Background,
	}
}

func (b *box) Render(_ tux.BuildContext, ctx tux.RenderContext) error {
	if b.width <= 0 || b.height <= 0 {
		return nil
	}

	for row := range b.height {
		for column := range b.width {
			ctx.Paint(b.row+row, b.column+column, tux.Cell{
				Ch:         ' ',
				Width:      uint8(misc.RuneWidth(' ')),
				Foreground: b.foreground,
				Background: b.background,
			})
		}
	}

	row := 0
	column := 0
	for _, ch := range b.content {
		width := misc.RuneWidth(ch)
		if width == 0 {
			continue
		}

		if column+width > b.width {
			row++
			column = 0
		}
		if row >= b.height {
			break
		}
		if width > b.width {
			break
		}

		ctx.Paint(b.row+row, b.column+column, tux.Cell{
			Ch:         ch,
			Width:      uint8(width),
			Foreground: b.foreground,
			Background: b.background,
		})
		if width == 2 {
			ctx.Paint(b.row+row, b.column+column+1, tux.Cell{
				Width:      0,
				Foreground: b.foreground,
				Background: b.background,
			})
		}
		column += width
	}

	return nil
}
