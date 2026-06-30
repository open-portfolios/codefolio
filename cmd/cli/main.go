package main

import "github.com/grindlemire/go-tui"

func main() {
	app, err := tui.NewApp(tui.WithRootComponent(Hello()))
	if err != nil {
		panic(err)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
