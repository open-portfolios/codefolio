package components

import (
	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
)

type PromptProps struct {
	State *state.State[builtin.TextareaState]
	OnKey func(input.KeyEvent) bool
}
type Prompt struct {
	state *state.State[builtin.TextareaState]
	onKey func(input.KeyEvent) bool
}

func NewPrompt(ctx renderer.Context, props PromptProps, children ...renderer.Component) *Prompt {
	return &Prompt{state: props.State, onKey: props.OnKey}
}
