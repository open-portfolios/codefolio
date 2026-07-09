//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/cmd/cli/tui"
	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra"
)

func InitModel(cfg *conf.Global) (*tui.Model, func(), error) {
	wire.Build(domain.Set, infra.Set, tui.NewModel)
	return nil, nil, nil
}
