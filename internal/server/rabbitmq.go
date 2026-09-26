package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/itsping999/greeting/internal/pkg/messagingx"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/protocol"
	"github.com/itsping999/greeting/internal/service"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQServer struct {
	mu                          sync.Mutex
	stopped                     bool
	conn                        *amqp.Connection
	addr                        string
	exchange, routingKey, queue string
	svc                         *service.GreeterService
	log                         *log.Helper
}

func NewRabbitMQServer(c *RabbitMQConfig, logger log.Logger, svc *service.GreeterService) *RabbitMQServer {
	if c == nil || !c.Enabled {
		return nil
	}

	exchange, routingKey, queue := c.Exchange, c.RoutingKey, c.Queue
	if exchange == "" {
		exchange = protocol.RabbitExchange
	}
	if routingKey == "" {
		routingKey = protocol.RabbitRequestKey
	}
	if queue == "" {
		queue = protocol.RabbitRequestQueue
	}
	return &RabbitMQServer{exchange: exchange, routingKey: routingKey, queue: queue, addr: c.Addr, svc: svc, log: log.NewHelper(logger)}
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
	if err := ch.ExchangeDeclare(s.exchange, "direct", true, false, false, false, nil); err != nil {
		return s.transportError(err)
	}
	queue, err := ch.QueueDeclare(s.queue, true, false, false, false, nil)
	if err != nil {
		return s.transportError(err)
	}
	if err := ch.QueueBind(queue.Name, s.routingKey, s.exchange, false, nil); err != nil {
		return s.transportError(err)
	}
	if err := ch.Qos(16, 0, false); err != nil {
		return s.transportError(err)
	}
	deliveries, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return s.transportError(err)
	}
	s.log.Infof("[RabbitMQ] subscribed to %s", s.routingKey)
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
			err = ch.PublishWithContext(requestCtx, s.exchange, protocol.RabbitReplyPrefix+in.RequestID, false, false, amqp.Publishing{ContentType: "application/json", CorrelationId: in.RequestID, Body: payload})
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

type RabbitMQConfig = messagingx.RabbitMQConfig
