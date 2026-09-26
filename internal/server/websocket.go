package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gorilla/websocket"
	"github.com/itsping999/greeting/internal/conf"
	"github.com/itsping999/greeting/internal/protocol"
	"github.com/itsping999/greeting/internal/service"
)

type wsClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}
type WebsocketServer struct {
	mu         sync.Mutex
	clients    map[*wsClient]struct{}
	stopped    bool
	httpServer *http.Server
	listener   net.Listener
	workers    sync.WaitGroup
	c          *conf.Server_Websocket
	svc        *service.GreeterService
	log        *log.Helper
}

func NewWebsocketServer(c *conf.Server_Websocket, logger log.Logger, svc *service.GreeterService) *WebsocketServer {
	s := &WebsocketServer{c: c, svc: svc, clients: make(map[*wsClient]struct{}), log: log.NewHelper(logger)}
	mux := http.NewServeMux()
	mux.HandleFunc(c.Path, s.handle)
	s.httpServer = &http.Server{Handler: mux, ReadHeaderTimeout: conf.Timeout(c.Timeout)}
	return s
}
func (s *WebsocketServer) Start(context.Context) error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil
	}
	lis, err := net.Listen("tcp", s.c.Addr)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	s.listener = lis
	s.mu.Unlock()
	s.log.Infof("[WebSocket] listening on %s%s", lis.Addr(), s.c.Path)
	err = s.httpServer.Serve(lis)
	if errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
func (s *WebsocketServer) handle(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &wsClient{conn: conn}
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		conn.Close()
		return
	}
	s.clients[client] = struct{}{}
	s.workers.Add(1)
	s.mu.Unlock()
	defer s.workers.Done()
	defer func() { conn.Close(); s.mu.Lock(); delete(s.clients, client); s.mu.Unlock() }()
	conn.SetReadLimit(4096)
	timeout := conf.Timeout(s.c.Timeout)
	for {
		conn.SetReadDeadline(time.Now().Add(timeout))
		var in protocol.ChatMessage
		if err := conn.ReadJSON(&in); err != nil {
			return
		}
		out := s.svc.OnChatMessage(r.Context(), in)
		if out.Error != "" {
			if err := s.send(client, out); err != nil {
				return
			}
			continue
		}
		s.mu.Lock()
		clients := make([]*wsClient, 0, len(s.clients))
		for c := range s.clients {
			clients = append(clients, c)
		}
		s.mu.Unlock()
		for _, c := range clients {
			if err := s.send(c, out); err != nil {
				c.conn.Close()
			}
		}
	}
}
func (s *WebsocketServer) send(c *wsClient, out protocol.ChatMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.SetWriteDeadline(time.Now().Add(conf.Timeout(s.c.Timeout)))
	return c.conn.WriteJSON(out)
}
func (s *WebsocketServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	s.stopped = true
	for c := range s.clients {
		c.conn.Close()
	}
	s.mu.Unlock()
	err := s.httpServer.Shutdown(ctx)
	done := make(chan struct{})
	go func() { s.workers.Wait(); close(done) }()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
