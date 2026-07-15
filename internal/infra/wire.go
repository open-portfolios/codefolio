package infra

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/llm"
	"github.com/open-portfolios/codefolio/internal/infra/session"
	"github.com/open-portfolios/codefolio/internal/infra/tools"
)

func NewExecutorFactory() domain.ExecutorFactory {
	return NewExecutor
}

var Set = wire.NewSet(
	llm.NewDriver,
	session.New,
	tools.NewRegistry,
	NewExecutorFactory,
	NewEnvironmentService,
)
