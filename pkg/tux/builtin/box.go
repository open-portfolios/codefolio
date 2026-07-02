package builtin

import "github.com/open-portfolios/codefolio/pkg/tux"

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
	foreground tux.Color
	background tux.Color
}

type BoxProps struct {
	Row        int
	Column     int
	Width      int
	Height     int
	Content    string
	Foreground tux.Color
	Background tux.Color
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
			index := row*b.width + column
			ch := byte(' ')
			if index < len(b.content) {
				ch = b.content[index]
			}

			ctx.Paint(b.row+row, b.column+column, tux.Cell{
				Ch:         ch,
				Foreground: b.foreground,
				Background: b.background,
			})
		}
	}

	return nil
}
