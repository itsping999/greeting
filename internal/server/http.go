package server

import (
	nethttp "net/http"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/api/helloworld/v1"
	"github.com/itsping999/greeting/internal/pkg/metricsx"
	"github.com/itsping999/greeting/internal/pkg/tracingx"
	"github.com/itsping999/greeting/internal/service"

	"github.com/go-kratos/kratos/v2/transport/http"
)

type HTTPConfig struct {
	Enabled bool          `json:"enabled" yaml:"enabled"`
	Network string        `json:"network" yaml:"network"`
	Addr    string        `json:"addr" yaml:"addr"`
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *HTTPConfig,
	middlewareCfg *MiddlewareConfig,
	tracingCfg *tracingx.Config,
	logger log.Logger,
	greeter *service.GreeterService,
	system *service.SystemService,
) *http.Server {
	if c == nil || !c.Enabled {
		return nil
	}
	srv := newHTTPServer(c, middlewareCfg, tracingCfg, logger, system)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	return srv
}

// newHTTPServer builds shared transport routes independently of business services.
func newHTTPServer(
	c *HTTPConfig,
	middlewareCfg *MiddlewareConfig,
	tracingCfg *tracingx.Config,
	logger log.Logger,
	system *service.SystemService,
) *http.Server {
	helper := log.NewHelper(log.With(logger, "module", "server/http"))
	if c == nil || !c.Enabled {
		helper.Info("http server disabled, skip initialization")
		return nil
	}
	var opts = []http.ServerOption{
		http.Middleware(commonMiddlewares(logger, tracingCfg, middlewareCfg)...),
	}
	if c.Network != "" {
		opts = append(opts, http.Network(c.Network))
	}
	if c.Addr != "" {
		opts = append(opts, http.Address(c.Addr))
	}
	if c.Timeout > 0 {
		opts = append(opts, http.Timeout(c.Timeout))
	}
	srv := http.NewServer(opts...)
	r := srv.Route("/")
	r.GET("/healthz", func(ctx http.Context) error {
		return ctx.Result(200, system.Liveness())
	})
	r.GET("/readyz", func(ctx http.Context) error {
		ready, payload := system.ReadinessStatus()
		statusCode := nethttp.StatusOK
		if !ready {
			statusCode = nethttp.StatusServiceUnavailable
		}
		return ctx.Result(statusCode, payload)
	})
	r.GET("/meta", func(ctx http.Context) error {
		return ctx.Result(200, system.Metadata())
	})
	r.GET("/metrics", func(ctx http.Context) error {
		handler := metricsx.Handler()
		if handler == nil {
			return ctx.Result(nethttp.StatusServiceUnavailable, map[string]any{
				"status":  "error",
				"message": "metrics exporter is not initialized",
			})
		}
		handler.ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	})

	return srv
}
