package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	v1 "github.com/itsping999/greeting/api/helloworld/v1"
	"github.com/itsping999/greeting/internal/protocol"
)

func (s *GreeterService) OnMQTTMessage(ctx context.Context, in *protocol.GreetingRequest) *protocol.GreetingReply {
	return s.greetMessage(ctx, in)
}
func (s *GreeterService) greetMessage(ctx context.Context, in *protocol.GreetingRequest) *protocol.GreetingReply {
	out := &protocol.GreetingReply{RequestID: in.RequestID}
	greeting, err := s.SayHello(ctx, &v1.HelloRequest{Name: in.Name})
	if err != nil {
		if errors.IsBadRequest(err) {
			out.Error = "name must contain 1 to 256 characters"
		} else {
			out.Error = "greeting processing failed"
		}
		return out
	}
	out.Message = greeting.Message
	out.ID = greeting.Id
	s.log.WithContext(ctx).Infof("greeting processed request_id=%s id=%d", in.RequestID, out.ID)
	return out
}
