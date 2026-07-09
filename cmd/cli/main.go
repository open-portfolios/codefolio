package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	dotenv "github.com/joho/godotenv"

	"github.com/open-portfolios/codefolio/internal/conf"
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

	model, cleanup, err := InitModel(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init error: %v\n", err)
		os.Exit(1)
	}
	if cleanup != nil {
		defer cleanup()
	}

	p := tea.NewProgram(model)
	model.Program = p

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
