package main

import (
	"log"
	"strings"

	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/builtin"
	"github.com/open-portfolios/codefolio/pkg/tux/misc"
)

func main() {
	inputState := tux.NewState(builtin.InputState{
		Content:   "",
		CursorPos: 0,
	})

	root := &demo{
		input: inputState,
	}

	app := tux.NewApp(
		root,
		tux.WithSize(6, 40),
		tux.WithFrameRate(30),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

type demo struct {
	tux.Composite

	input *tux.State[builtin.InputState]
}

func (d *demo) Build(ctx tux.BuildContext) tux.Component {
	return builtin.Container(
		builtin.ContainerProps{},
		builtin.Box(builtin.BoxProps{
			Row:     0,
			Column:  0,
			Width:   40,
			Height:  1,
			Content: "tux testcli",
		}),
		builtin.Box(builtin.BoxProps{
			Row:     1,
			Column:  0,
			Width:   40,
			Height:  1,
			Content: "type text/中文; arrows move; Ctrl+C exits",
		}),
		builtin.Box(builtin.BoxProps{
			Row:     3,
			Column:  0,
			Width:   40,
			Height:  1,
			Content: "input:",
		}),
		builtin.Input(builtin.InputProps{
			Row:        4,
			Column:     0,
			Width:      24,
			State:      d.input,
			AutoFocus:  true,
			Background: misc.ColorGreen,
		}),
		builtin.Box(builtin.BoxProps{
			Row:     5,
			Column:  0,
			Width:   40,
			Height:  1,
			Content: strings.Repeat("-", 40),
		}),
	)
}
