package components

import (
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
	"github.com/cylixlee/tux/style"
)

type PromptProps struct {
	State   *state.State[builtin.TextareaState]
	OnKey   func(input.KeyEvent) bool
	Model   string
	WorkDir string
}
type Prompt struct {
	state   *state.State[builtin.TextareaState]
	onKey   func(input.KeyEvent) bool
	model   string
	workDir string
}

func NewPrompt(ctx renderer.Context, props PromptProps, children ...renderer.Component) *Prompt {
	return &Prompt{state: props.State, onKey: props.OnKey, model: props.Model, workDir: props.WorkDir}
}

func (p *Prompt) Render(ctx renderer.Context) *renderer.Element {
	width, _ := ctx.Size()
	if width > 120 {
		width -= 42
	}
	textareaWidth := max(width-7, 1) // Main inset, rail, and composer padding.
	inputHeight := builtin.TextareaPreferredHeight(p.state.Value().Value, textareaWidth, 1, 8)
	input := builtin.CreateRow(ctx, builtin.RowProps{Key: "composer-input", Height: inputHeight, Bg: Theme.BackgroundElement},
		builtin.CreateTextarea(ctx, builtin.TextareaProps{
			Key:         "editor",
			State:       p.state,
			Placeholder: "Ask anything...",
			MinHeight:   1,
			MaxHeight:   8,
			AutoFocus:   true,
			Fg:          Theme.Text,
			Bg:          Theme.BackgroundElement,
			OnKey:       p.onKey,
		}),
	)
	meta := builtin.CreateRow(ctx, builtin.RowProps{Key: "composer-meta", Gap: 1, Bg: Theme.BackgroundElement},
		builtin.CreateText(ctx, builtin.TextProps{Text: "Build", Fg: Theme.Primary, Bg: Theme.BackgroundElement}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "·", Fg: Theme.TextMuted, Bg: Theme.BackgroundElement}),
		builtin.CreateText(ctx, builtin.TextProps{Text: p.model, Fg: Theme.Text, Bg: Theme.BackgroundElement}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "DeepSeek", Fg: Theme.TextMuted, Bg: Theme.BackgroundElement}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "·", Fg: Theme.TextMuted, Bg: Theme.BackgroundElement}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "High", Fg: Theme.Warning, Bg: Theme.BackgroundElement, Style: style.Bold}),
	)
	composer := builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "composer", Padding: 1, Gap: 1, Bg: Theme.BackgroundElement}, input, meta)
	status := builtin.CreateRow(ctx, builtin.RowProps{Key: "prompt-status", Gap: 1, Bg: Theme.Background},
		builtin.CreateText(ctx, builtin.TextProps{Text: p.workDir, Fg: Theme.TextMuted, Bg: Theme.Background, Style: style.Dim}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "·", Fg: Theme.TextMuted, Bg: Theme.Background, Style: style.Dim}),
		builtin.CreateText(ctx, builtin.TextProps{Text: "Ctrl+P commands", Fg: Theme.TextMuted, Bg: Theme.Background}),
	)
	separator := builtin.CreateText(ctx, builtin.TextProps{Key: "prompt-separator", Text: "", Bg: Theme.Background})
	bottomSpace := builtin.CreateText(ctx, builtin.TextProps{Key: "prompt-bottom-space", Text: "", Bg: Theme.Background})
	return builtin.CreateColumn(ctx, builtin.ColumnProps{Key: "prompt"}, NewRail(ctx, RailProps{}, composer), separator, status, bottomSpace)
}
