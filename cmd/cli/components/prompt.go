package components

import (
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
)

type PromptProps struct {
	State    *state.State[builtin.TextareaState]
	OnKey    func(input.KeyEvent) bool
	Model    string
	WorkDir  string
	Disabled bool
	Focus    bool
}
type Prompt struct {
	state    *state.State[builtin.TextareaState]
	onKey    func(input.KeyEvent) bool
	model    string
	workDir  string
	disabled bool
	focus    bool
}

func NewPrompt(ctx renderer.Context, props PromptProps, children ...renderer.Component) *Prompt {
	return &Prompt{state: props.State, onKey: props.OnKey, model: props.Model, workDir: props.WorkDir, disabled: props.Disabled, focus: props.Focus}
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
	buildFg, highFg, highAttrs := Theme.Primary, Theme.Warning, style.Bold
	if p.disabled {
		metaFg, buildFg, highFg = Theme.TextMuted, Theme.TextMuted, Theme.TextMuted
		metaAttrs, highAttrs = style.Dim, style.Dim
	}
	meta := builtin.CreateRow(ctx, builtin.RowProps{Key: "composer-meta", Gap: 1, Bg: Theme.BackgroundElement},
		builtin.CreateText(ctx, builtin.TextProps{Text: "Build", Fg: buildFg, Bg: Theme.BackgroundElement, Style: metaAttrs}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "·", Fg: Theme.TextMuted, Bg: Theme.BackgroundElement}),
		builtin.CreateText(ctx, builtin.TextProps{Text: p.model, Fg: metaFg, Bg: Theme.BackgroundElement, Style: metaAttrs}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "DeepSeek", Fg: Theme.TextMuted, Bg: Theme.BackgroundElement, Style: metaAttrs}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "·", Fg: Theme.TextMuted, Bg: Theme.BackgroundElement}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "High", Fg: highFg, Bg: Theme.BackgroundElement, Style: highAttrs}),
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
		builtin.CreateText(ctx, builtin.TextProps{Text: "Ctrl+P commands", Fg: Theme.TextMuted, Bg: Theme.Background}),
	)
	separator := builtin.CreateText(ctx, builtin.TextProps{Key: "prompt-separator", Text: "", Bg: Theme.Background})
	bottomSpace := builtin.CreateText(ctx, builtin.TextProps{Key: "prompt-bottom-space", Text: "", Bg: Theme.Background})
	return builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "prompt"}, NewRail(ctx, RailProps{Disabled: p.disabled}, composer), separator, status, bottomSpace)
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
