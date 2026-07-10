package tools

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/internal/infra/tools/file"
)

func ProvideRegistry(dt DefaultTools) *Registry {
	return dt.Registry
}

var Set = wire.NewSet(
	CreateDefaultTools,
	file.NewStateCache,
	ProvideRegistry,
)
