package components

import (
	"testing"

	"github.com/cylixlee/tux/builtin"
	"github.com/cylixlee/tux/input"
	"github.com/cylixlee/tux/renderer"
	"github.com/cylixlee/tux/state"
)

func TestComposerClickFocusesTextarea(t *testing.T) {
	var focused *renderer.Element
	prompt := NewPrompt(renderer.Context{}, PromptProps{State: state.New(builtin.TextareaState{PreferredColumn: -1})})
	root := prompt.Render(renderer.Context{SizeFn: func() (int, int) { return 80, 24 }, FocusFn: func(element *renderer.Element) { focused = element }})
	composer := root.Children()[0].Children()[0]
	if !composer.HandleMouse(input.KeyEvent{Type: input.EventMouse, Mouse: input.MouseEvent{Action: input.MousePress, Button: input.MouseLeft}}) {
		t.Fatal("composer click should be handled")
	}
	if focused == nil || focused.Tag() != "textarea" {
		t.Fatalf("focused element = %#v, want textarea", focused)
	}
}

func TestDisabledComposerShowsStreamingStateAndCannotFocus(t *testing.T) {
	var focused *renderer.Element
	prompt := NewPrompt(renderer.Context{}, PromptProps{State: state.New(builtin.TextareaState{PreferredColumn: -1}), Disabled: true})
	root := prompt.Render(renderer.Context{SizeFn: func() (int, int) { return 80, 24 }, FocusFn: func(element *renderer.Element) { focused = element }})
	composer := root.Children()[0].Children()[0]
	textarea := composer.Children()[0].Children()[0]
	props := renderer.Props[builtin.TextareaProps](textarea)
	if !props.Disabled || props.Placeholder != "LLM Streaming..." {
		t.Fatalf("textarea props = %#v, want disabled streaming placeholder", props)
	}
	if composer.HandleMouse(input.KeyEvent{Type: input.EventMouse, Mouse: input.MouseEvent{Action: input.MousePress, Button: input.MouseLeft}}) {
		t.Fatal("disabled composer must not handle focus clicks")
	}
	if focused != nil {
		t.Fatalf("disabled composer focused %#v", focused)
	}
}

func TestPromptShowsOnlyProfileAndModel(t *testing.T) {
	prompt := NewPrompt(renderer.Context{}, PromptProps{State: state.New(builtin.TextareaState{}), Model: "test-model", Profile: "build"})
	root := prompt.Render(renderer.Context{SizeFn: func() (int, int) { return 80, 24 }})
	meta := root.Children()[0].Children()[0].Children()[1]
	if got := len(meta.Children()); got != 3 {
		t.Fatalf("composer metadata item count = %d, want profile, separator, and model", got)
	}
}
