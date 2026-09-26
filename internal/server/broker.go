package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/protocol"
	"github.com/tx7do/kratos-transport/broker"
)

// BrokerServer subscribes after connecting, so subscription failures are returned.
type BrokerServer struct {
	mu                       sync.Mutex
	broker                   broker.Broker
	subscriber               broker.Subscriber
	stopped                  bool
	connected                bool
	topic, replyPrefix, name string
	opts                     []broker.SubscribeOption
	handle                   func(context.Context, *protocol.GreetingRequest) *protocol.GreetingReply
	log                      *log.Helper
}

func (s *BrokerServer) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return nil
	}
	if err := s.broker.Init(); err != nil {
		return fmt.Errorf("%s initialization failed", s.name)
	}
	if err := s.broker.Connect(); err != nil {
		return fmt.Errorf("%s connection failed", s.name)
	}
	s.connected = true
	sub, err := s.broker.Subscribe(s.topic, func(ctx context.Context, evt broker.Event) error {
		in, ok := evt.Message().Body.(*protocol.GreetingRequest)
		if !ok || in == nil || !protocol.ValidRequestID(in.RequestID) {
			return fmt.Errorf("invalid greeting request or request_id")
		}
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return s.broker.Publish(ctx, s.replyPrefix+in.RequestID, s.handle(ctx, in))
	}, func() broker.Any { return &protocol.GreetingRequest{} }, s.opts...)
	if err != nil {
		s.broker.Disconnect()
		s.connected = false
		return fmt.Errorf("%s subscription failed", s.name)
	}
	s.subscriber = sub
	s.log.Infof("[%s] subscribed to %s", s.name, s.topic)
	return nil
}
func (s *BrokerServer) Stop(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopped = true
	if s.subscriber != nil {
		s.subscriber.Unsubscribe(true)
		s.subscriber = nil
	}
	if s.connected {
		s.connected = false
		return s.broker.Disconnect()
	}
	return nil
}
