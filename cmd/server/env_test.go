package main

import (
	"github.com/itsping999/greeting/internal/conf"
	"strings"
	"testing"
)

func TestDeploymentEnvironmentOverrides(t *testing.T) {
	t.Setenv("APP_SERVER_HTTP_ADDR", "127.0.0.1:18000")
	t.Setenv("APP_SERVER_HTTP_ENABLED", "true")
	t.Setenv("APP_DATA_REMOTE_GRPC_GREETER_ENABLED", "true")
	t.Setenv("APP_DATA_REMOTE_GRPC_GREETER_TARGET", "greeter:9000")
	t.Setenv("GREETING_KCP_KEY", "test-key")
	t.Setenv("GREETING_KCP_SALT", "test-salt")
	var c conf.Bootstrap
	if err := applyEnvOverrides(&c); err != nil {
		t.Fatal(err)
	}
	if !c.Server.HTTP.Enabled || c.Server.HTTP.Addr != "127.0.0.1:18000" || !c.Data.RemoteGRPC.Greeter.Enabled || c.Data.RemoteGRPC.Greeter.Target != "greeter:9000" || c.Server.KCP.Key != "test-key" || c.Server.KCP.Salt != "test-salt" {
		t.Fatal("environment override not applied")
	}
}
func TestInvalidEnvironmentBooleanDoesNotExposeValue(t *testing.T) {
	const value = "invalid-private-value"
	t.Setenv("APP_SERVER_HTTP_ENABLED", value)
	var c conf.Bootstrap
	err := applyEnvOverrides(&c)
	if err == nil || !strings.Contains(err.Error(), "APP_SERVER_HTTP_ENABLED") || strings.Contains(err.Error(), value) {
		t.Fatalf("unexpected error: %v", err)
	}
}
