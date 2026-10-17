package tracingx

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
)

func TestInitDoesNotLogEndpointCredentials(t *testing.T) {
	previousProvider := otel.GetTracerProvider()
	previousPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})

	var logs bytes.Buffer
	endpoint := "https://collector.example.invalid/v1/traces?token=example-secret"
	cleanup, err := Init(&Config{
		Enabled:  true,
		Exporter: "stdout",
		Endpoint: endpoint,
	}, "greeting-test", "test", log.NewStdLogger(&logs))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	if !strings.Contains(logs.String(), "tracing enabled") {
		t.Fatal("expected tracing initialization log")
	}
	if strings.Contains(logs.String(), endpoint) || strings.Contains(logs.String(), "example-secret") {
		t.Fatal("tracing initialization logged endpoint credentials")
	}
}
