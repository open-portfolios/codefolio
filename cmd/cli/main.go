package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cylixlee/tux/app"
	"github.com/cylixlee/tux/terminal"
	dotenv "github.com/joho/godotenv"

	"github.com/open-portfolios/codefolio/cmd/cli/components"
	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/infra/approval"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
)

func init() {
	if err := dotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
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
	approvalCh := make(chan *approval.Request, 16)
	root, cleanup, err := InitApp(cfg, askUserCh, approvalCh)
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
	runtime := app.New(app.AppConfig{Root: root, Terminal: term, Background: components.Theme.Background})
	root.AttachApp(runtime)
	runErr := runtime.Run(context.Background())
	root.Shutdown()
	term.Restore()
	fmt.Fprintf(os.Stdout, "\n%s\n\n", components.CodefolioBannerText())
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", runErr)
		os.Exit(1)
	}
}
