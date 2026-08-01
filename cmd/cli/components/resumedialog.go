package components

import (
	"fmt"

	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
	"github.com/open-portfolios/codefolio/internal/domain"
)

type ResumeDialogProps struct {
	Open      *state.State[bool]
	Sessions  []domain.SessionInfo
	Selected  int
	OnMove    func(int)
	OnConfirm func()
	OnCancel  func()
}

type ResumeDialog struct{ props ResumeDialogProps }

func NewResumeDialog(ctx renderer.Context, props ResumeDialogProps, children ...renderer.Component) *ResumeDialog {
	return &ResumeDialog{props: props}
}

func (d *ResumeDialog) Render(ctx renderer.Context) *renderer.Element {
	if d.props.Open == nil || !d.props.Open.Value() {
		hidden := &renderer.Element{}
		hidden.SetHidden(true)
		return hidden
	}
	items := make([]string, len(d.props.Sessions))
	for i, session := range d.props.Sessions {
		items[i] = fmt.Sprintf("%s  %s  (%d messages)", session.ID, session.Title, session.MessageCount)
	}
	keys := builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "resume-keys", Text: "Choose a previous session to restore.", Width: 64, Fg: Theme.Text})
	keys.SetHandleKeyFn(func(event input.KeyEvent) bool {
		switch {
		case event.Key == input.KeyArrowUp || event.Rune == 'k':
			d.props.OnMove(-1)
		case event.Key == input.KeyArrowDown || event.Rune == 'j':
			d.props.OnMove(1)
		case event.Key == input.KeyEnter:
			d.props.OnConfirm()
		case event.Key == input.KeyEscape:
			d.props.OnCancel()
		default:
			return false
		}
		return true
	})
	content := builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "resume-content", Padding: 1, Gap: 1, Bg: Theme.BackgroundPanel},
		keys,
		builtin.CreateList(ctx, builtin.ListProps{Key: "resume-sessions", Items: items, Selected: d.props.Selected, Height: min(max(len(items), 1), 8), Fg: Theme.Primary}),
		builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "resume-help", Text: "Up/Down select  ·  Enter restore  ·  Esc cancel", Width: 64, Fg: Theme.TextMuted, Style: style.Dim}),
	)
	return builtin.CreateModal(ctx, builtin.ModalProps{Key: "resume-modal", Open: d.props.Open, Backdrop: "dim", CloseOnEscape: false, Title: " Resume session ", Border: builtin.BorderRounded, BorderColor: Theme.Accent}, content)
}
