package llm

import (
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	anthropicoption "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/openai/openai-go/v3"
	openaioption "github.com/openai/openai-go/v3/option"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/pkg/llm"
	anthropicmessages "github.com/open-portfolios/codefolio/pkg/llm/anthropic/messages"
	openaichat "github.com/open-portfolios/codefolio/pkg/llm/openai/chat"
)

func NewDriver(cfg *conf.Global) (llm.Driver, error) {
	switch cfg.Protocol {
	case "anthropic":
		client := anthropic.NewClient(
			anthropicoption.WithBaseURL(cfg.BaseURL),
			anthropicoption.WithAPIKey(cfg.APIKey),
		)
		return anthropicmessages.NewDriver(&client), nil
	case "openai":
		client := openai.NewClient(
			openaioption.WithBaseURL(cfg.BaseURL),
			openaioption.WithAPIKey(cfg.APIKey),
		)
		return openaichat.NewCompletionsDriver(&client), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}
}
