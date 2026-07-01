package components

import tui "github.com/grindlemire/go-tui"

type app struct{}

func App() *app { return new(app) }

func (a *app) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
	}
}

templ (a *app) Render() {
    <div class="w-full h-full justify-center items-center flex-col gap-1">
        <p>Hello, TUI!</p>
        <p class="font-dim">Press Esc to quit</p>
    </div>
}
