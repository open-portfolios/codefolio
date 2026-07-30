package infra

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra/llm"
	"github.com/open-portfolios/codefolio/internal/infra/mcp"
	"github.com/open-portfolios/codefolio/internal/infra/session"
	"github.com/open-portfolios/codefolio/internal/infra/tools"
)

func NewExecutorFactory(authorizer domain.Authorizer) domain.ExecutorFactory {
	return func(registry domain.ToolRegistry, workDir string) domain.Executor {
		return NewExecutor(registry, authorizer, workDir)
	}
}

var Set = wire.NewSet(
	llm.NewDriver,
	mcp.NewManager,
	session.New,
	tools.NewRegistry,
	wire.Bind(new(domain.ToolRegistry), new(*tools.Registry)),
	wire.Bind(new(domain.ToolRegistrar), new(*tools.Registry)),
	NewExecutorFactory,
	NewEnvironmentService,
)
