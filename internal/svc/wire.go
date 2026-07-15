package svc

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/internal/svc/prompt"
)

var Set = wire.NewSet(NewAgent, prompt.NewPromptService)
