package builtin

import "github.com/open-portfolios/codefolio/pkg/tux"

var (
	_ tux.Component = (*container)(nil)
)

type container struct {
	tux.Atomic

	children []tux.Component
}

type ContainerProps struct{}

func Container(props ContainerProps, children ...tux.Component) tux.Component {
	return &container{children: children}
}

func (c *container) Render(b tux.BuildContext, r tux.RenderContext) error {
	for _, child := range c.children {
		for {
			if artifact := child.Build(b); artifact != nil {
				if artifact == child {
					panic("composition cycle is not allowed")
				}
				child = artifact
				continue
			}
			break
		}
		if err := child.Render(b, r); err != nil {
			return err
		}
	}
	return nil
}
