package service

import (
	"context"
	"io"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/biz"
	"github.com/itsping999/greeting/internal/data"
	"github.com/itsping999/greeting/internal/protocol"
)

func TestMessageRepliesPreserveCorrelationAndErrors(t *testing.T) {
	logger := log.NewStdLogger(io.Discard)
	store := data.NewData(logger)
	svc := NewGreeterService(logger, biz.NewGreeterUsecase(data.NewGreeterRepo(store)))
	for _, handler := range []func(context.Context, *protocol.GreetingRequest) *protocol.GreetingReply{svc.OnMQTTMessage, svc.OnRabbitMQMessage} {
		out := handler(context.Background(), &protocol.GreetingRequest{RequestID: "test-1", Name: " World "})
		if out.RequestID != "test-1" || out.Message != "Hello World" || out.ID == 0 || out.Error != "" {
			t.Fatalf("greeting reply %+v", out)
		}
		invalid := handler(context.Background(), &protocol.GreetingRequest{RequestID: "test-2", Name: " "})
		if invalid.RequestID != "test-2" || invalid.Error == "" || invalid.ID != 0 {
			t.Fatalf("invalid-name reply %+v", invalid)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		cancelled := handler(ctx, &protocol.GreetingRequest{RequestID: "test-3", Name: "World"})
		if cancelled.Error != "greeting processing failed" {
			t.Fatalf("cancelled reply %+v", cancelled)
		}
	}
}
