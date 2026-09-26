package server

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"net"
	"sync"
	"time"
)

type UDPServer struct {
	mu      sync.Mutex
	conn    *net.UDPConn
	stopped bool
	c       *UDPConfig
	log     *log.Helper
}

func NewUDPServer(c *UDPConfig, logger log.Logger) *UDPServer {
	if c == nil || !c.Enabled {
		return nil
	}

	return &UDPServer{c: c, log: log.NewHelper(logger)}
}
func (s *UDPServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil
	}
	addr, err := net.ResolveUDPAddr("udp", s.c.Addr)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	s.conn = conn
	s.mu.Unlock()
	s.log.Infof("[UDP] listening on %s", conn.LocalAddr())
	buf := make([]byte, 65535)
	timeout := Timeout(s.c.Timeout)
	for {
		conn.SetReadDeadline(time.Now().Add(timeout))
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return err
		}
		conn.SetWriteDeadline(time.Now().Add(timeout))
		if _, err := conn.WriteToUDP(buf[:n], remote); err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
	}
}
func (s *UDPServer) Stop(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopped = true
	if s.conn == nil {
		return nil
	}
	err := s.conn.Close()
	if errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}
