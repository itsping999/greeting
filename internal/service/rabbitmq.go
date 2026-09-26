package service

import (
	"context"
	"github.com/itsping999/greeting/internal/protocol"
)

func (s *GreeterService) OnRabbitMQMessage(ctx context.Context, in *protocol.GreetingRequest) *protocol.GreetingReply {
	return s.greetMessage(ctx, in)
}
