package main

import (
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	tea "charm.land/bubbletea/v2"
	dotenv "github.com/joho/godotenv"

	"github.com/open-portfolios/codefolio/cmd/cli/tui"
	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/pkg/llm/anthropic/messages"
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

	client := anthropic.NewClient(
		option.WithBaseURL(cfg.BaseURL),
		option.WithAPIKey(cfg.APIKey),
	)
	driver := messages.NewDriver(&client)

	model := tui.NewModel(cfg, driver)

	p := tea.NewProgram(model)
	model.Program = p

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
