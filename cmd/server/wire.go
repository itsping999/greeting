//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/biz"
	"github.com/itsping999/greeting/internal/data"
	provider "github.com/itsping999/greeting/internal/pkg"
	"github.com/itsping999/greeting/internal/pkg/tracingx"
	"github.com/itsping999/greeting/internal/server"
	"github.com/itsping999/greeting/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*server.Config, *data.Config, *tracingx.Config, log.Logger) (*kratos.App, func(), error) {
	panic(
		wire.Build(
			wire.FieldsOf(new(*server.Config), "HTTP", "GRPC", "Middleware", "WebSocket", "MQTT", "RabbitMQ", "TCP", "UDP", "KCP"),
			wire.FieldsOf(new(*data.Config), "Postgres", "Redis", "MongoDB", "Discovery", "RemoteGRPC"),
			server.ProviderSet,
			provider.ProviderSet,
			data.ProviderSet,
			biz.ProviderSet,
			service.ProviderSet,
			newApp,
		),
	)
}
