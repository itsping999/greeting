package service

import (
	"context"
	v1 "github.com/itsping999/greeting/api/helloworld/v1"
	"github.com/itsping999/greeting/internal/protocol"
)

func (s *GreeterService) OnChatMessage(ctx context.Context, in protocol.ChatMessage) protocol.ChatMessage {
	if in.Type != 1 {
		return protocol.ChatMessage{Type: in.Type, Error: "unsupported message type"}
	}
	out, err := s.SayHello(ctx, &v1.HelloRequest{Name: in.Name})
	if err != nil {
		return protocol.ChatMessage{Type: 1, Error: "name must contain 1 to 256 characters"}
	}
	return protocol.ChatMessage{Type: 1, Message: out.Message}
}
