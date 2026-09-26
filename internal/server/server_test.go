package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gorilla/websocket"
	"github.com/itsping999/greeting/internal/biz"
	"github.com/itsping999/greeting/internal/conf"
	"github.com/itsping999/greeting/internal/data"
	"github.com/itsping999/greeting/internal/protocol"
	"github.com/itsping999/greeting/internal/service"
	"google.golang.org/protobuf/types/known/durationpb"
)

func testGreeter() *service.GreeterService {
	logger := log.NewStdLogger(io.Discard)
	store := data.NewData(logger)
	return service.NewGreeterService(logger, biz.NewGreeterUsecase(data.NewGreeterRepo(store)))
}
func TestHTTPGreetingAndErrors(t *testing.T) {
	srv := NewHTTPServer(&conf.Server_HTTP{}, log.NewStdLogger(io.Discard), testGreeter())
	for _, tt := range []struct {
		path     string
		status   int
		contains string
	}{
		{"/helloworld/World", 200, "Hello World"}, {"/greetings/1", 200, "Hello World"},
		{"/greetings/999", 404, "GREETING_NOT_FOUND"}, {"/greetings/0", 400, ""},
		{"/helloworld/%20", 400, "INVALID_NAME"}, {"/healthz", 200, "ok"},
	} {
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if rr.Code != tt.status {
			t.Errorf("%s status %d: %s", tt.path, rr.Code, rr.Body.String())
		}
		if tt.contains != "" && !strings.Contains(rr.Body.String(), tt.contains) {
			t.Errorf("%s response %s", tt.path, rr.Body.String())
		}
	}
}
func awaitStream(t *testing.T, s *streamServer) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		lis := s.listener
		s.mu.Unlock()
		if lis != nil {
			return lis.Addr().String()
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("listener not ready")
	return ""
}
func TestTCPTimeoutAndShutdown(t *testing.T) {
	srv := NewTCPServer(&conf.Server_TCP{Addr: "127.0.0.1:0", Timeout: durationpb.New(100 * time.Millisecond)}, log.NewStdLogger(io.Discard))
	result := make(chan error, 1)
	go func() { result <- srv.Start(context.Background()) }()
	conn, err := net.Dial("tcp", awaitStream(t, srv.streamServer))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	reply := make([]byte, 5)
	if _, err := io.ReadFull(conn, reply); err != nil || string(reply) != "hello" {
		t.Fatalf("echo %q %v", reply, err)
	}
	if _, err := conn.Read(reply); err == nil {
		t.Fatal("idle connection was not closed")
	}
	active, err := net.Dial("tcp", awaitStream(t, srv.streamServer))
	if err != nil {
		t.Fatal(err)
	}
	defer active.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	active.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := active.Read(reply); err == nil {
		t.Fatal("active connection survived shutdown")
	}
	if err := srv.Stop(ctx); err != nil {
		t.Fatal(err)
	}
}
func TestStopAfterFailedStart(t *testing.T) {
	logger := log.NewStdLogger(io.Discard)
	for name, srv := range map[string]interface {
		Start(context.Context) error
		Stop(context.Context) error
	}{
		"tcp":       NewTCPServer(&conf.Server_TCP{Addr: "invalid"}, logger),
		"udp":       NewUDPServer(&conf.Server_UDP{Addr: "invalid"}, logger),
		"kcp":       NewKCPServer(&conf.Server_KCP{Addr: "invalid"}, logger),
		"websocket": NewWebsocketServer(&conf.Server_Websocket{Addr: "invalid", Path: "/ws"}, logger, testGreeter()),
	} {
		t.Run(name, func(t *testing.T) {
			if err := srv.Start(context.Background()); err == nil {
				t.Fatal("start unexpectedly succeeded")
			}
			if err := srv.Stop(context.Background()); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestWebsocketBroadcastAndShutdown(t *testing.T) {
	srv := NewWebsocketServer(&conf.Server_Websocket{Addr: "127.0.0.1:0", Path: "/ws", Timeout: durationpb.New(time.Second)}, log.NewStdLogger(io.Discard), testGreeter())
	result := make(chan error, 1)
	go func() { result <- srv.Start(context.Background()) }()
	var addr string
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); {
		srv.mu.Lock()
		lis := srv.listener
		srv.mu.Unlock()
		if lis != nil {
			addr = lis.Addr().String()
			break
		}
		time.Sleep(time.Millisecond)
	}
	if addr == "" {
		t.Fatal("listener not ready")
	}
	clients := make([]*websocket.Conn, 0, 2)
	for i := 0; i < 2; i++ {
		c, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		c.SetReadDeadline(time.Now().Add(2 * time.Second))
		clients = append(clients, c)
	}
	// A response confirms that the second connection is registered.
	if err := clients[1].WriteJSON(protocol.ChatMessage{Type: 0}); err != nil {
		t.Fatal(err)
	}
	var invalid protocol.ChatMessage
	if err := clients[1].ReadJSON(&invalid); err != nil || invalid.Error == "" {
		t.Fatalf("invalid type reply %+v %v", invalid, err)
	}
	if err := clients[0].WriteJSON(protocol.ChatMessage{Type: 1, Name: "World"}); err != nil {
		t.Fatal(err)
	}
	for _, c := range clients {
		var out protocol.ChatMessage
		if err := c.ReadJSON(&out); err != nil || out.Message != "Hello World" {
			t.Fatalf("broadcast %+v %v", out, err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if _, _, err := clients[0].ReadMessage(); err == nil {
		t.Fatal("WebSocket survived shutdown")
	}
}
func TestDisabledProtocolsNeedNoConfig(t *testing.T) {
	servers, err := NewServers(&conf.Server{Http: &conf.Server_HTTP{Enabled: true, Addr: "127.0.0.1:0"}}, log.NewStdLogger(io.Discard), testGreeter())
	if err != nil || len(servers.All()) != 1 {
		t.Fatalf("servers: %v %v", fmt.Sprint(servers), err)
	}
}
