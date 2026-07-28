package components

import "github.com/cylixlee/tux/renderer"

type HeaderProps struct {
	Model   string
	WorkDir string
}
type Header struct {
	model   string
	workDir string
}

func NewHeader(ctx renderer.Context, props HeaderProps, children ...renderer.Component) *Header {
	return &Header{model: props.Model, workDir: props.WorkDir}
}
