package components

import (
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
)

type PromptProps struct {
	State      *state.State[builtin.TextareaState]
	OnKey      func(input.KeyEvent) bool
	Model      string
	Profile    string
	WorkDir    string
	Disabled   bool
	Focus      bool
	Completion CommandCompletion
}

type CommandCompletion struct {
	Open       bool
	Candidates []CommandCandidate
	Selected   int
}

type CommandCandidate struct{ Name, Description string }
type Prompt struct {
	state      *state.State[builtin.TextareaState]
	onKey      func(input.KeyEvent) bool
	model      string
	profile    string
	workDir    string
	disabled   bool
	focus      bool
	completion CommandCompletion
}

func NewPrompt(ctx renderer.Context, props PromptProps, children ...renderer.Component) *Prompt {
	return &Prompt{state: props.State, onKey: props.OnKey, model: props.Model, profile: props.Profile, workDir: props.WorkDir, disabled: props.Disabled, focus: props.Focus, completion: props.Completion}
}

func (p *Prompt) Render(ctx renderer.Context) *renderer.Element {
	width, _ := ctx.Size()
	if width > 120 {
		width -= 42
	}
	textareaWidth := max(width-7, 1) // Main inset, rail, and composer padding.
	inputHeight := builtin.TextareaPreferredHeight(p.state.Value().Value, textareaWidth, 1, 8)
	textarea := builtin.CreateTextarea(ctx, builtin.TextareaProps{
		Key:         "editor",
		State:       p.state,
		Placeholder: placeholder(p.disabled),
		MinHeight:   1,
		MaxHeight:   8,
		AutoFocus:   false,
		Disabled:    p.disabled,
		Focus:       p.focus,
		Fg:          promptTextColor(p.disabled),
		Bg:          Theme.BackgroundElement,
		OnKey:       p.onKey,
	})
	if !p.disabled {
		textarea.SetHandleMouseFn(func(ev input.KeyEvent) bool {
			if ev.Mouse.Action != input.MousePress || ev.Mouse.Button != input.MouseLeft {
				return false
			}
			ctx.Focus(textarea)
			return true
		})
	}
	inputRow := builtin.CreateRow(ctx, builtin.RowProps{Key: "composer-input", Height: inputHeight, Bg: Theme.BackgroundElement}, textarea)
	metaFg, metaAttrs := Theme.Text, style.Attr(0)
	profileFg := ProfileColor(p.profile)
	if p.disabled {
		metaFg, profileFg = Theme.TextMuted, Theme.TextMuted
		metaAttrs = style.Dim
	}
	meta := builtin.CreateRow(ctx, builtin.RowProps{Key: "composer-meta", Gap: 1, Bg: Theme.BackgroundElement},
		builtin.CreateText(ctx, builtin.TextProps{Text: profileLabel(p.profile), Fg: profileFg, Bg: Theme.BackgroundElement, Style: metaAttrs}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "·", Fg: Theme.TextMuted, Bg: Theme.BackgroundElement}),
		builtin.CreateText(ctx, builtin.TextProps{Text: p.model, Fg: metaFg, Bg: Theme.BackgroundElement, Style: metaAttrs}),
	)
	composer := builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "composer", Padding: 1, Gap: 1, Bg: Theme.BackgroundElement}, inputRow, meta)
	if !p.disabled {
		composer.SetHandleMouseFn(func(ev input.KeyEvent) bool {
			if ev.Mouse.Action != input.MousePress || ev.Mouse.Button != input.MouseLeft {
				return false
			}
			ctx.Focus(textarea)
			return true
		})
	}
	status := builtin.CreateRow(ctx, builtin.RowProps{Key: "prompt-status", Gap: 1, Bg: Theme.Background},
		builtin.CreateText(ctx, builtin.TextProps{Text: p.workDir, Fg: Theme.TextMuted, Bg: Theme.Background, Style: style.Dim}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "·", Fg: Theme.TextMuted, Bg: Theme.Background, Style: style.Dim}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "Tab switch mode", Fg: Theme.TextMuted, Bg: Theme.Background}),
	)
	separator := builtin.CreateText(ctx, builtin.TextProps{Key: "prompt-separator", Text: "", Bg: Theme.Background})
	bottomSpace := builtin.CreateText(ctx, builtin.TextProps{Key: "prompt-bottom-space", Text: "", Bg: Theme.Background})
	root := builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "prompt"}, NewRail(ctx, RailProps{Disabled: p.disabled, Color: ProfileColor(p.profile)}, composer), separator, status, bottomSpace)
	if p.completion.Open {
		root.SetChildren(append(root.Children(), p.completionPopup(ctx, width, inputHeight)))
	}
	return root
}

func (p *Prompt) completionPopup(ctx renderer.Context, width, inputHeight int) *renderer.Element {
	_, height := ctx.Size()
	popupHeight := len(p.completion.Candidates)
	// Shell places Prompt at the terminal bottom: composer, separator, status,
	// and one trailing spacer occupy inputHeight+7 rows in total.
	y := height - (inputHeight + 7) - popupHeight
	popupWidth := min(max(width-8, 28), 72)
	items := make([]renderer.Component, 0, len(p.completion.Candidates))
	for index, candidate := range p.completion.Candidates {
		prefix, fg, bg := "  ", Theme.TextMuted, Theme.BackgroundPanel
		if index == p.completion.Selected {
			prefix, fg, bg = "> ", Theme.Primary, Theme.BackgroundElement
		}
		items = append(items, builtin.CreateText(ctx, builtin.TextProps{Key: "command-popup-" + candidate.Name, Text: prefix + "/" + candidate.Name + "  " + candidate.Description, Fg: fg, Bg: bg}))
	}
	content := builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "command-popup-content", Bg: Theme.BackgroundPanel}, items...)
	return builtin.CreatePopup(ctx, builtin.PopupProps{Key: "command-popup", X: 3, Y: y, Width: popupWidth, Height: popupHeight, Bg: Theme.BackgroundPanel}, content)
}

func placeholder(disabled bool) string {
	if disabled {
		return "LLM Streaming..."
	}
	return "Ask anything..."
}

func promptTextColor(disabled bool) style.Color {
	if disabled {
		return Theme.TextMuted
	}
	return Theme.Text
}

func profileLabel(profile string) string {
	if profile == "plan" {
		return "Plan"
	}
	return "Build"
}
