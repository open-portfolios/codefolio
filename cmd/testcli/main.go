package main

import (
	"log"
	"strings"

	"github.com/open-portfolios/codefolio/pkg/tux"
	"github.com/open-portfolios/codefolio/pkg/tux/builtin"
	"github.com/open-portfolios/codefolio/pkg/tux/misc"
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
					runes := []rune(root.content)
					root.content = string(append(runes[:root.cursor-1], runes[root.cursor:]...))
					root.cursor--
					app.MarkDirty()
				}
			case tux.KeyDelete:
				runes := []rune(root.content)
				if root.cursor < len(runes) {
					root.content = string(append(runes[:root.cursor], runes[root.cursor+1:]...))
					app.MarkDirty()
				}
			case tux.KeyLeft:
				if root.cursor > 0 {
					root.cursor--
					app.MarkDirty()
				}
			case tux.KeyRight:
				if root.cursor < len([]rune(root.content)) {
					root.cursor++
					app.MarkDirty()
				}
			case tux.KeyRune:
				if event.Rune >= 32 {
					root.content = insertRune(root.content, root.cursor, event.Rune)
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

func insertRune(s string, index int, r rune) string {
	runes := []rune(s)
	if index < 0 {
		index = 0
	}
	if index > len(runes) {
		index = len(runes)
	}

	next := make([]rune, 0, len(runes)+1)
	next = append(next, runes[:index]...)
	next = append(next, r)
	next = append(next, runes[index:]...)
	return string(next)
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
			Content:    d.content,
			Cursor:     d.cursor,
			Focused:    true,
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
