package main

import (
	"fmt"
	"github.com/itsping999/greeting/internal/conf"
	"os"
	"strconv"
)

// applyEnvOverrides applies the deployment-facing keys without exposing values in errors.
func applyEnvOverrides(c *conf.Bootstrap) error {
	strings := map[string]*string{
		"APP_SERVER_HTTP_ADDR":                &c.Server.HTTP.Addr,
		"APP_SERVER_GRPC_ADDR":                &c.Server.GRPC.Addr,
		"APP_DATA_POSTGRES_SOURCE":            &c.Data.Postgres.Source,
		"APP_DATA_REDIS_ADDR":                 &c.Data.Redis.Addr,
		"APP_DATA_REDIS_PASSWORD":             &c.Data.Redis.Password,
		"APP_DATA_MONGODB_URI":                &c.Data.MongoDB.URI,
		"APP_DATA_REMOTE_GRPC_GREETER_TARGET": &c.Data.RemoteGRPC.Greeter.Target,
		"APP_TRACING_ENDPOINT":                &c.Tracing.Endpoint,
		"APP_LOGGER_LEVEL":                    &c.Logger.Level,
		"GREETING_KCP_KEY":                    &c.Server.KCP.Key,
		"GREETING_KCP_SALT":                   &c.Server.KCP.Salt,
	}
	for key, target := range strings {
		if value, ok := os.LookupEnv(key); ok {
			*target = value
		}
	}
	booleans := map[string]*bool{
		"APP_SERVER_HTTP_ENABLED":                 &c.Server.HTTP.Enabled,
		"APP_SERVER_GRPC_ENABLED":                 &c.Server.GRPC.Enabled,
		"APP_SERVER_MIDDLEWARE_RATELIMIT_ENABLED": &c.Server.Middleware.RateLimitEnabled,
		"APP_DATA_DISCOVERY_KUBERNETES_ENABLED":   &c.Data.Discovery.Kubernetes.Enabled,
		"APP_DATA_POSTGRES_ENABLED":               &c.Data.Postgres.Enabled,
		"APP_DATA_REDIS_ENABLED":                  &c.Data.Redis.Enabled,
		"APP_DATA_MONGODB_ENABLED":                &c.Data.MongoDB.Enabled,
		"APP_DATA_REMOTE_GRPC_GREETER_ENABLED":    &c.Data.RemoteGRPC.Greeter.Enabled,
		"APP_TRACING_ENABLED":                     &c.Tracing.Enabled,
	}
	for key, target := range booleans {
		if value, ok := os.LookupEnv(key); ok {
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("%s must be a boolean", key)
			}
			*target = parsed
		}
	}
	return nil
}
