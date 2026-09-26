package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/conf"
	"github.com/itsping999/greeting/internal/server"
	_ "go.uber.org/automaxprocs"
)

var (
	Name     = "greeting"
	Version  = "dev"
	flagconf string
	id, _    = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs/config.yaml", "configuration file or directory")
}
func newApp(logger log.Logger, servers *server.Servers) *kratos.App {
	return kratos.New(kratos.ID(id), kratos.Name(Name), kratos.Version(Version), kratos.Logger(logger), kratos.Server(servers.All()...))
}
func run() error {
	logger := log.With(log.NewStdLogger(os.Stdout), "ts", log.DefaultTimestamp, "caller", log.DefaultCaller, "service.id", id, "service.name", Name, "service.version", Version)
	c := config.New(config.WithSource(file.NewSource(flagconf)))
	defer c.Close()
	if err := c.Load(); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		return fmt.Errorf("decode configuration: %w", err)
	}
	if bc.Server != nil && bc.Server.Kcp != nil {
		if value := os.Getenv("GREETING_KCP_KEY"); value != "" {
			bc.Server.Kcp.Key = value
		}
		if value := os.Getenv("GREETING_KCP_SALT"); value != "" {
			bc.Server.Kcp.Salt = value
		}
	}
	if err := conf.Validate(bc.Server); err != nil {
		return err
	}
	app, err := wireApp(bc.Server, logger)
	if err != nil {
		return err
	}
	return app.Run()
}
func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
