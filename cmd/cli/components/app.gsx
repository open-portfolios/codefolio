package components

import tui "github.com/grindlemire/go-tui"

type app struct {
	body     *tui.Ref
	composer *tui.Ref
	textarea *tui.Ref

	input *tui.State[string]
}

func App() *app {
	return &app{
		body:     tui.NewRef(),
		composer: tui.NewRef(),
		textarea: tui.NewRef(),
		input:    tui.NewState(""),
	}
}

func (a *app) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.On(tui.KeyCtrlC, a.onCtrlC),
	}
}

func (a *app) HandleMouse(m tui.MouseEvent) bool {
	return tui.HandleClicks(m,
		tui.Click(a.composer, func() { a.onComposerClick(m) }),
		tui.Click(a.body, func() { a.onBodyClick(m) }),
	)
}

templ (a *app) Render() {
	<div ref={a.body} class="w-full h-full bg-black flex">
		@MainPanel() {
			@Messages()
			@Composer(a.composer) {
				<textarea ref={a.textarea} value={a.input} placeholder="Ask anything..." autoFocus={true} />
			}
		}
		if a.shouldShowSidebar(app) {
			@Sidebar()
		}
	</div>
}

func (a *app) shouldShowSidebar(t *tui.App) bool {
	w, _ := t.Size()
	return w >= 120
}

func (a *app) onCtrlC(ke tui.KeyEvent) {
	textarea := tui.RefComponent[*tui.TextArea](a.textarea)
	if textarea.IsFocused() && a.input.Get() != "" {
		a.input.Set("")
		return
	}
	ke.App().Stop()
}

func (a *app) onComposerClick(m tui.MouseEvent) {
	textarea := tui.RefComponent[*tui.TextArea](a.textarea)
	if !textarea.IsFocused() {
		m.App().FocusNext()
	}
}

func (a *app) onBodyClick(m tui.MouseEvent) { m.App().BlurFocused() }
