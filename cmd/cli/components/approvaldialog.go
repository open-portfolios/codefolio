package components

import (
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
	"github.com/open-portfolios/codefolio/cmd/cli/controller"
)

type ApprovalDialogProps struct {
	Open           *state.State[bool]
	Approval       controller.ApprovalState
	OnAllowOnce    func()
	OnAllowSession func()
	OnDeny         func()
}

type ApprovalDialog struct{ props ApprovalDialogProps }

func NewApprovalDialog(ctx renderer.Context, props ApprovalDialogProps, children ...renderer.Component) *ApprovalDialog {
	return &ApprovalDialog{props: props}
}

func (d *ApprovalDialog) Render(ctx renderer.Context) *renderer.Element {
	toolName, summary, detail := "Tool", "Waiting for approval", "This action needs your approval."
	if request := d.props.Approval.Request; request != nil {
		toolName, summary, detail = request.ToolName, request.Summary, request.Detail
	}
	once := builtin.CreateButton(ctx, builtin.ButtonProps{Key: "approval-once", Text: "Allow once", OnClick: d.props.OnAllowOnce})
	session := builtin.CreateButton(ctx, builtin.ButtonProps{Key: "approval-session", Text: "Allow this session", OnClick: d.props.OnAllowSession})
	deny := builtin.CreateButton(ctx, builtin.ButtonProps{Key: "approval-deny", Text: "Deny", OnClick: d.props.OnDeny})
	for _, button := range []*renderer.Element{once, session, deny} {
		previous := button.HandleKeyFn
		button.SetHandleKeyFn(func(event input.KeyEvent) bool {
			return approvalKey(event, d) || previous(event)
		})
	}
	content := builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "approval-content", Padding: 1, Gap: 1, Bg: Theme.BackgroundPanel},
		builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "approval-action", Text: toolName + " wants to run:", Width: 56, Fg: Theme.Text}),
		builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "approval-summary", Text: summary, Width: 56, Fg: Theme.Primary, Bg: Theme.BackgroundElement}),
		builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "approval-detail", Text: detail, Width: 56, Fg: Theme.TextMuted}),
		builtin.CreateRow(ctx, builtin.RowProps{Key: "approval-actions", Gap: 1}, once, session, deny),
		builtin.CreateTextBlock(ctx, builtin.TextBlockProps{Key: "approval-help", Text: "a allow once  ·  s allow session  ·  d / Esc deny", Width: 56, Fg: Theme.TextMuted, Style: style.Dim}),
	)
	return builtin.CreateModal(ctx, builtin.ModalProps{Key: "approval-modal", Open: d.props.Open, Backdrop: "dim", CloseOnEscape: false, Title: " Permission required ", Border: builtin.BorderRounded, BorderColor: Theme.Warning}, content)
}

func approvalKey(event input.KeyEvent, d *ApprovalDialog) bool {
	switch {
	case event.Rune == 'a' && event.Modifiers == 0:
		d.props.OnAllowOnce()
	case event.Rune == 's' && event.Modifiers == 0:
		d.props.OnAllowSession()
	case event.Rune == 'd' && event.Modifiers == 0:
		d.props.OnDeny()
	case event.Key == input.KeyEscape || (event.Rune == 'c' && event.Modifiers == input.ModCtrl):
		d.props.OnDeny()
	default:
		return false
	}
	return true
}
