package main

import (
	"log"
	"strings"

	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/builtin"
)

func main() {
	root := &demo{
		content: "hello input",
		cursor:  5,
	}
	app := tux.NewApp(
		root,
		tux.WithSize(6, 40),
		tux.WithFrameRate(30),
		tux.WithInput(func(app *tux.App, event tux.InputEvent) {
			switch event.Key {
			case tux.KeyBackspace:
				if root.cursor > 0 {
					root.content = root.content[:root.cursor-1] + root.content[root.cursor:]
					root.cursor--
					app.MarkDirty()
				}
			case tux.KeyLeft:
				if root.cursor > 0 {
					root.cursor--
					app.MarkDirty()
				}
			case tux.KeyRight:
				if root.cursor < len(root.content) {
					root.cursor++
					app.MarkDirty()
				}
			case tux.KeyByte:
				if event.Byte >= 32 && event.Byte <= 126 {
					root.content = root.content[:root.cursor] + string(event.Byte) + root.content[root.cursor:]
					root.cursor++
					app.MarkDirty()
				}
			}
		}),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

type demo struct {
	tux.Composite

	content string
	cursor  int
}

func (d *demo) Build(tux.BuildContext) tux.Component {
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
			Content: "type text; arrows move; Esc exits",
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
			Content:    d.content,
			Cursor:     d.cursor,
			Focused:    true,
			Background: tux.ColorGreen,
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
