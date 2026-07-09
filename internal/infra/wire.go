package infra

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/internal/infra/llm"
)

var Set = wire.NewSet(llm.NewDriver)
