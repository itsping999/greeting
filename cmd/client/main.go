package main

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"encoding/json"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/websocket"
	v1 "github.com/itsping999/greeting/api/helloworld/v1"
	"github.com/itsping999/greeting/internal/protocol"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tx7do/kratos-transport/broker"
	"github.com/tx7do/kratos-transport/broker/mqtt"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	kind := flag.String("protocol", "http", "http, grpc, websocket, mqtt, rabbitmq, tcp, udp or kcp")
	addr := flag.String("addr", "", "server address; default depends on protocol")
	name := flag.String("name", "World", "greeting name")
	timeout := flag.Duration("timeout", 10*time.Second, "overall client timeout")
	flag.Parse()
	if *addr == "" {
		*addr = map[string]string{"http": "http://127.0.0.1:8000", "grpc": "127.0.0.1:9000", "websocket": "ws://127.0.0.1:8001/ws", "mqtt": "tcp://127.0.0.1:1883", "rabbitmq": "amqp://guest:guest@127.0.0.1:5672/", "tcp": "127.0.0.1:9001", "udp": "127.0.0.1:9002", "kcp": "127.0.0.1:9003"}[*kind]
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	if err := run(ctx, *kind, *addr, *name, *timeout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(ctx context.Context, kind, addr, name string, timeout time.Duration) error {
	name = strings.TrimSpace(name)
	switch kind {
	case "http":
		conn, err := kratoshttp.NewClient(ctx, kratoshttp.WithEndpoint(addr), kratoshttp.WithTimeout(timeout))
		if err != nil {
			return err
		}
		defer conn.Close()
		client := v1.NewGreeterHTTPClient(conn)
		created, err := client.SayHello(ctx, &v1.HelloRequest{Name: name})
		if err != nil {
			return err
		}
		found, err := client.GetGreeting(ctx, &v1.GetGreetingRequest{Id: created.Id})
		if err != nil {
			return err
		}
		if created.Id != found.Id || created.Message != found.Message {
			return fmt.Errorf("saved greeting does not match")
		}
		fmt.Printf("HTTP: %s (saved and retrieved id=%d)\n", found.Message, found.Id)
		return nil
	case "grpc":
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		defer conn.Close()
		client := v1.NewGreeterClient(conn)
		created, err := client.SayHello(ctx, &v1.HelloRequest{Name: name})
		if err != nil {
			return err
		}
		found, err := client.GetGreeting(ctx, &v1.GetGreetingRequest{Id: created.Id})
		if err != nil {
			return err
		}
		if created.Id != found.Id || created.Message != found.Message {
			return fmt.Errorf("saved greeting does not match")
		}
		fmt.Printf("gRPC: %s (saved and retrieved id=%d)\n", found.Message, found.Id)
		return nil
	case "websocket":
		return chat(ctx, addr, name)
	case "mqtt":
		return messaging(ctx, kind, addr, name, timeout)
	case "rabbitmq":
		return rabbitMessaging(ctx, addr, name, timeout)
	case "tcp", "udp", "kcp":
		return echo(ctx, kind, addr, name)
	default:
		return fmt.Errorf("unknown protocol %q", kind)
	}
}
func chat(ctx context.Context, addr, name string) error {
	d := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	first, _, err := d.DialContext(ctx, addr, nil)
	if err != nil {
		return err
	}
	defer first.Close()
	second, _, err := d.DialContext(ctx, addr, nil)
	if err != nil {
		return err
	}
	defer second.Close()
	deadline, _ := ctx.Deadline()
	first.SetWriteDeadline(deadline)
	first.SetReadDeadline(deadline)
	second.SetReadDeadline(deadline)
	// Verify the second session has entered its read loop before testing broadcast.
	second.SetWriteDeadline(deadline)
	if err := second.WriteJSON(protocol.ChatMessage{Type: 0}); err != nil {
		return err
	}
	var ready protocol.ChatMessage
	if err := second.ReadJSON(&ready); err != nil {
		return err
	}
	if ready.Error == "" {
		return fmt.Errorf("expected unsupported-type response")
	}
	if err := first.WriteJSON(protocol.ChatMessage{Type: 1, Name: name}); err != nil {
		return err
	}
	for _, c := range []*websocket.Conn{first, second} {
		var out protocol.ChatMessage
		if err := c.ReadJSON(&out); err != nil {
			return err
		}
		if out.Error != "" {
			return fmt.Errorf("%s", out.Error)
		}
		if out.Type != 1 || out.Message != "Hello "+name {
			return fmt.Errorf("unexpected broadcast: %+v", out)
		}
	}
	fmt.Printf("WebSocket: Hello %s (received by both clients)\n", name)
	return nil
}
func messaging(ctx context.Context, kind, addr, name string, timeout time.Duration) error {
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return err
	}
	id := hex.EncodeToString(random)
	var b broker.Broker
	topic, reply := protocol.MQTTRequestTopic, protocol.MQTTReplyPrefix+id
	var opts []broker.SubscribeOption
	b = mqtt.NewBroker(broker.WithAddress(addr), broker.WithCodec("json"), mqtt.WithConnectTimeout(timeout), mqtt.WithWriteTimeout(timeout))
	opts = []broker.SubscribeOption{mqtt.WithSubscribeQos(1)}
	if err := b.Init(); err != nil {
		return fmt.Errorf("%s initialization failed", kind)
	}
	if err := b.Connect(); err != nil {
		return fmt.Errorf("%s connection failed", kind)
	}
	defer b.Disconnect()
	received := make(chan protocol.GreetingReply, 1)
	sub, err := b.Subscribe(reply, func(_ context.Context, event broker.Event) error {
		response, ok := event.Message().Body.(*protocol.GreetingReply)
		if !ok || response == nil {
			return fmt.Errorf("invalid greeting reply")
		}
		select {
		case received <- *response:
		default:
		}
		return nil
	}, func() broker.Any { return &protocol.GreetingReply{} }, opts...)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe(true)
	var publishOpts []broker.PublishOption
	if kind == "mqtt" {
		publishOpts = append(publishOpts, mqtt.WithPublishQos(1))
	}
	if err := b.Publish(ctx, topic, &protocol.GreetingRequest{RequestID: id, Name: name}, publishOpts...); err != nil {
		return err
	}
	select {
	case out := <-received:
		if out.RequestID != id {
			return fmt.Errorf("reply request_id does not match")
		}
		if out.Error != "" {
			return fmt.Errorf("%s", out.Error)
		}
		if out.Message != "Hello "+name || out.ID <= 0 {
			return fmt.Errorf("unexpected greeting reply")
		}
		fmt.Printf("%s: %s (request/reply matched)\n", kind, out.Message)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%s reply: %w", kind, ctx.Err())
	}
}
func echo(ctx context.Context, kind, addr, name string) error {
	var conn net.Conn
	var err error
	if kind == "kcp" {
		secret, salt := os.Getenv("GREETING_KCP_KEY"), os.Getenv("GREETING_KCP_SALT")
		if secret == "" || salt == "" {
			return fmt.Errorf("set GREETING_KCP_KEY and GREETING_KCP_SALT")
		}
		key := pbkdf2.Key([]byte(secret), []byte(salt), 4096, 32, sha1.New)
		block, err := kcp.NewAESBlockCrypt(key)
		if err != nil {
			return err
		}
		conn, err = kcp.DialWithOptions(addr, block, 10, 3)
		if err != nil {
			return err
		}
	} else {
		conn, err = (&net.Dialer{}).DialContext(ctx, kind, addr)
		if err != nil {
			return err
		}
	}
	defer conn.Close()
	deadline, _ := ctx.Deadline()
	conn.SetDeadline(deadline)
	input := []byte("Hello " + name)
	if len(input) > 60000 {
		return fmt.Errorf("echo payload is too large")
	}
	if _, err := conn.Write(input); err != nil {
		return err
	}
	out := make([]byte, len(input))
	if kind == "udp" {
		n, err := conn.Read(out)
		if err != nil {
			return err
		}
		if n != len(input) {
			return fmt.Errorf("truncated UDP reply")
		}
	} else {
		if _, err := io.ReadFull(conn, out); err != nil {
			return err
		}
	}
	if string(out) != string(input) {
		return fmt.Errorf("echo reply does not match")
	}
	fmt.Printf("%s: %s (echo matched)\n", kind, out)
	return nil
}

func rabbitMessaging(ctx context.Context, addr, name string, timeout time.Duration) error {
	conn, err := amqp.DialConfig(addr, amqp.Config{Dial: func(network, address string) (net.Conn, error) {
		conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, network, address)
		if err == nil {
			deadline, _ := ctx.Deadline()
			conn.SetDeadline(deadline)
		}
		return conn, err
	}})
	if err != nil {
		return fmt.Errorf("RabbitMQ connection failed")
	}
	deadline, _ := ctx.Deadline()
	stop := context.AfterFunc(ctx, func() { conn.CloseDeadline(time.Now()) })
	defer stop()
	defer conn.CloseDeadline(deadline)
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(protocol.RabbitExchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	queue, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return err
	}
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return err
	}
	id := hex.EncodeToString(random)
	reply := protocol.RabbitReplyPrefix + id
	if err := ch.QueueBind(queue.Name, reply, protocol.RabbitExchange, false, nil); err != nil {
		return err
	}
	deliveries, err := ch.Consume(queue.Name, "", true, true, false, false, nil)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(protocol.GreetingRequest{RequestID: id, Name: name})
	if err != nil {
		return err
	}
	if err := ch.PublishWithContext(ctx, protocol.RabbitExchange, protocol.RabbitRequestKey, false, false, amqp.Publishing{ContentType: "application/json", DeliveryMode: amqp.Persistent, CorrelationId: id, Body: payload}); err != nil {
		return err
	}
	select {
	case delivery, ok := <-deliveries:
		if !ok {
			return fmt.Errorf("RabbitMQ reply consumer closed")
		}
		var out protocol.GreetingReply
		if err := json.Unmarshal(delivery.Body, &out); err != nil {
			return err
		}
		if out.RequestID != id {
			return fmt.Errorf("reply request_id does not match")
		}
		if out.Error != "" {
			return fmt.Errorf("%s", out.Error)
		}
		if out.Message != "Hello "+name || out.ID <= 0 {
			return fmt.Errorf("unexpected greeting reply")
		}
		fmt.Printf("RabbitMQ: %s (request/reply matched)\n", out.Message)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("RabbitMQ reply: %w", ctx.Err())
	}
}
