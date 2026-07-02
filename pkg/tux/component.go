package tux

import "github.com/open-portfolios/codefolio/pkg/stdx"

var (
	_ Component = stdx.Zero[Atomic]()
	_ Component = stdx.Zero[Composite]()
)

type Component interface {
	Build(BuildContext) Component
	Render(BuildContext, RenderContext) error
}

type Atomic struct{}

func (Atomic) Build(BuildContext) Component { return nil }
func (Atomic) Render(BuildContext, RenderContext) error {
	panic("atomic component must implement Render")
}

type Composite struct{}

func (Composite) Build(BuildContext) Component             { panic("composite component must implement Build") }
func (Composite) Render(BuildContext, RenderContext) error { return nil }
