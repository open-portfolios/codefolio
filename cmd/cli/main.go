package main

import (
	"github.com/grindlemire/go-tui"
	"github.com/open-portfolios/codefolio/cmd/cli/components"
)

func main() {
	app, err := tui.NewApp(tui.WithRootComponent(components.App()))
	if err != nil {
		panic(err)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
