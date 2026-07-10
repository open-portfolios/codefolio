//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/cmd/cli/tui"
	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra"
	"github.com/open-portfolios/codefolio/internal/infra/tools"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
	"github.com/open-portfolios/codefolio/internal/svc"
)

func provideProtocol(cfg *conf.Global) string {
	return cfg.Protocol
}

func askUserChannel() chan askuser.Request {
	return make(chan askuser.Request, 1)
}

func InitModel(cfg *conf.Global) (*tui.Model, func(), error) {
	wire.Build(domain.Set, infra.Set, tools.Set, svc.Set, provideProtocol, askUserChannel, tui.NewModel)
	return nil, nil, nil
}
