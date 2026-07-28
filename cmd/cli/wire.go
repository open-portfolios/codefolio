//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/infra"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
	"github.com/open-portfolios/codefolio/internal/svc"
)

func InitApp(cfg *conf.Global, askUserCh chan askuser.Request) (*App, func(), error) {
	wire.Build(infra.Set, svc.Set, NewApp)
	return nil, nil, nil
}
