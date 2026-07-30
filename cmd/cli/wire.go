//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
	"github.com/open-portfolios/codefolio/internal/infra"
	"github.com/open-portfolios/codefolio/internal/infra/approval"
	"github.com/open-portfolios/codefolio/internal/infra/tools/askuser"
	"github.com/open-portfolios/codefolio/internal/svc"
)

func InitApp(cfg *conf.Struct, askUserCh chan askuser.Request, approvalCh chan *approval.Request) (*App, func(), error) {
	wire.Build(infra.Set, svc.Set, approval.NewBroker, wire.Bind(new(domain.Authorizer), new(*approval.Broker)), NewApp)
	return nil, nil, nil
}
