//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/cmd/cli/tui"
	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/infra"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
	"github.com/open-portfolios/codefolio/internal/svc"
)

func InitModel(cfg *conf.Global, askUserCh chan askuser.Request) (*tui.Model, func(), error) {
	wire.Build(infra.Set, svc.Set, tui.NewModel)
	return nil, nil, nil
}
