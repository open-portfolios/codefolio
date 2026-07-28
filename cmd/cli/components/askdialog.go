package components

import (
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
	"github.com/open-portfolios/codefolio/cmd/cli/controller"
)

type AskDialogProps struct {
	Open       *state.State[bool]
	Ask        controller.AskState
	OnMove     func(int)
	OnConfirm  func()
	OnDefaults func()
}
type AskDialog struct{ props AskDialogProps }

func NewAskDialog(ctx renderer.Context, props AskDialogProps, children ...renderer.Component) *AskDialog {
	return &AskDialog{props: props}
}
func (d *AskDialog) Render(ctx renderer.Context) *renderer.Element {
	question := "Waiting for a question"
	items := []string(nil)
	if d.props.Ask.Request != nil && d.props.Ask.Question < len(d.props.Ask.Request.Questions) {
		q := d.props.Ask.Request.Questions[d.props.Ask.Question]
		question = q.Text
		items = make([]string, len(q.Options))
		for i, option := range q.Options {
			items[i] = option.Label
			if option.Description != "" {
				items[i] += " - " + option.Description
			}
		}
	}
	keys := builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "ask-keys", Text: question, Width: 40, Fg: primary})
	keys.SetHandleKeyFn(func(ev input.KeyEvent) bool {
		switch {
		case ev.Key == input.KeyArrowUp || ev.Rune == 'k':
			d.props.OnMove(-1)
		case ev.Key == input.KeyArrowDown || ev.Rune == 'j':
			d.props.OnMove(1)
		case ev.Key == input.KeyEnter:
			d.props.OnConfirm()
		case ev.Key == input.KeyEscape:
			d.props.OnDefaults()
		default:
			return false
		}
		return true
	})
	form := builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "ask-form", Padding: 1, Gap: 1}, keys, builtin.CreateList(ctx, builtin.ListProps{Key: "ask-options", Items: items, Selected: d.props.Ask.Selected, Height: min(max(len(items), 1), 4), Fg: primary}), builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "ask-help", Text: "Up/Down select | Enter confirm | Esc use defaults", Width: 40, Fg: muted, Style: style.Dim}))
	return builtin.CreateModal(ctx, builtin.ModalProps{Key: "ask-modal", Open: d.props.Open, Backdrop: "dim", CloseOnEscape: false, Title: " Agent question ", Border: builtin.BorderRounded, BorderColor: accent}, form)
}
