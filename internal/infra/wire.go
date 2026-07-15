package infra

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/internal/infra/llm"
	"github.com/open-portfolios/codefolio/internal/infra/session"
	"github.com/open-portfolios/codefolio/internal/infra/tools"
)

var Set = wire.NewSet(
	llm.NewDriver,
	session.New,
	tools.NewRegistry,
)
