package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/conf"
	"github.com/itsping999/greeting/internal/protocol"
	"github.com/itsping999/greeting/internal/service"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQServer struct {
	mu      sync.Mutex
	stopped bool
	conn    *amqp.Connection
	addr    string
	svc     *service.GreeterService
	log     *log.Helper
}

func NewRabbitMQServer(c *conf.Server_RabbitMQ, logger log.Logger, svc *service.GreeterService) *RabbitMQServer {
	return &RabbitMQServer{addr: c.Addr, svc: svc, log: log.NewHelper(logger)}
}
func (s *RabbitMQServer) isStopped() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.stopped }
func (s *RabbitMQServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil
	}
	conn, err := amqp.DialConfig(s.addr, amqp.Config{Dial: amqp.DefaultDial(5 * time.Second)})
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("RabbitMQ connection failed")
	}
	s.conn = conn
	s.mu.Unlock()
	ch, err := conn.Channel()
	if err != nil {
		return s.transportError(err)
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(protocol.RabbitExchange, "direct", true, false, false, false, nil); err != nil {
		return s.transportError(err)
	}
	queue, err := ch.QueueDeclare(protocol.RabbitRequestQueue, true, false, false, false, nil)
	if err != nil {
		return s.transportError(err)
	}
	if err := ch.QueueBind(queue.Name, protocol.RabbitRequestKey, protocol.RabbitExchange, false, nil); err != nil {
		return s.transportError(err)
	}
	if err := ch.Qos(16, 0, false); err != nil {
		return s.transportError(err)
	}
	deliveries, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return s.transportError(err)
	}
	s.log.Infof("[RabbitMQ] subscribed to %s", protocol.RabbitRequestKey)
	for delivery := range deliveries {
		var in protocol.GreetingRequest
		if err := json.Unmarshal(delivery.Body, &in); err != nil || !protocol.ValidRequestID(in.RequestID) {
			if err := delivery.Nack(false, false); err != nil {
				return s.transportError(err)
			}
			continue
		}
		requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		out := s.svc.OnRabbitMQMessage(requestCtx, &in)
		payload, err := json.Marshal(out)
		if err == nil {
			err = ch.PublishWithContext(requestCtx, protocol.RabbitExchange, protocol.RabbitReplyPrefix+in.RequestID, false, false, amqp.Publishing{ContentType: "application/json", CorrelationId: in.RequestID, Body: payload})
		}
		cancel()
		if err != nil {
			delivery.Nack(false, true)
			return s.transportError(err)
		}
		if err := delivery.Ack(false); err != nil {
			return s.transportError(err)
		}
	}
	if s.isStopped() {
		return nil
	}
	return fmt.Errorf("RabbitMQ consumer stopped unexpectedly")
}
func (s *RabbitMQServer) transportError(err error) error {
	if s.isStopped() {
		return nil
	}
	return fmt.Errorf("RabbitMQ transport failed: %w", err)
}
func (s *RabbitMQServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	s.stopped = true
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return nil
	}
	deadline := time.Now().Add(5 * time.Second)
	if limit, ok := ctx.Deadline(); ok && limit.Before(deadline) {
		deadline = limit
	}
	err := conn.CloseDeadline(deadline)
	if errors.Is(err, amqp.ErrClosed) {
		return nil
	}
	return err
}
