//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/itsping999/greeting/internal/biz"
	"github.com/itsping999/greeting/internal/conf"
	"github.com/itsping999/greeting/internal/data"
	"github.com/itsping999/greeting/internal/server"
	"github.com/itsping999/greeting/internal/service"
)

func wireApp(*conf.Server, log.Logger) (*kratos.App, error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
