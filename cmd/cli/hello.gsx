package main

import tui "github.com/grindlemire/go-tui"

type hello struct{}

func Hello() *hello { return new(hello) }

func (h *hello) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.On(tui.KeyEscape, func(e tui.KeyEvent) { e.App().Stop() }),
	}
}

templ (h *hello) Render() {
    <div class="h-full w-full justify-center items-center gap-1">
        <p>
            Hello, Go TUI!
        </p>
        <p class="font-dim">Press Esc to quit</p>
    </div>
}
