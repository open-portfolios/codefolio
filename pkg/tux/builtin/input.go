package builtin

import (
	"github.com/open-portfolios/codefolio/pkg/stdx"
	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/misc"
)

var (
	_ tux.Component = (*input)(nil)
)

type input struct {
	*box

	cursor  int
	focused bool
}

type InputProps struct {
	Row        int
	Column     int
	Width      int
	Content    string
	Cursor     int
	Focused    bool
	Foreground tux.Color
	Background tux.Color
}

func Input(props InputProps, children ...tux.Component) tux.Component {
	return &input{
		box: &box{
			row:        props.Row,
			column:     props.Column,
			width:      props.Width,
			height:     1,
			content:    props.Content,
			foreground: props.Foreground,
			background: props.Background,
		},
		cursor:  props.Cursor,
		focused: props.Focused,
	}
}

func (i *input) Render(build tux.BuildContext, ctx tux.RenderContext) error {
	if i.width <= 0 {
		return nil
	}

	if err := i.box.Render(build, ctx); err != nil {
		return err
	}

	if i.focused {
		ctx.QueueCursorMove(i.row, i.column+stdx.Clamp(misc.StringWidth(prefixRunes(i.content, i.cursor)), 0, i.width-1))
	}

	return nil
}

func prefixRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if n > len(runes) {
		n = len(runes)
	}
	return string(runes[:n])
}
