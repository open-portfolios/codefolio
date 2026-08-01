package svc

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/internal/svc/prompt"
)

var Set = wire.NewSet(NewContextManager, NewMemoryService, NewCommandRegistry, NewSessionService, NewAgent, prompt.NewPromptService)
