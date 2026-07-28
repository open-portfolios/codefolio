package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cylixlee/tux/app"
	"github.com/cylixlee/tux/terminal"
	dotenv "github.com/joho/godotenv"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
)

func init() {
	if err := dotenv.Load(); err != nil {
		panic(err)
	}
}

func main() {
	cfg, err := conf.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	askUserCh := make(chan askuser.Request, 1)
	model, cleanup, err := InitModel(cfg, askUserCh)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init error: %v\n", err)
		os.Exit(1)
	}
	if cleanup != nil {
		defer cleanup()
	}

	term, err := terminal.New(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "terminal error: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore()

	runtime := app.New(app.AppConfig{Root: model, Terminal: term})
	model.AttachApp(runtime)
	if err := runtime.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
