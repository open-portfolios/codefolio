package svc

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/internal/domain"
)

var Set = wire.NewSet(NewAgent, NewExecutorFactory)

func NewExecutorFactory() domain.ExecutorFactory {
	return NewExecutor
}
