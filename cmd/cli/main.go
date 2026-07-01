package main

import (
	"github.com/grindlemire/go-tui"
	"github.com/open-portfolios/codefolio/cmd/cli/components"
	"github.com/open-portfolios/codefolio/internal/conf"
)

func main() {
	if _, err := conf.Load(); err != nil {
		panic(err)
	}

	app, err := tui.NewApp(tui.WithRootComponent(components.App()))
	if err != nil {
		panic(err)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
