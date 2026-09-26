package server

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// streamServer tracks accepted connections as well as the listening socket.
type streamServer struct {
	mu          sync.Mutex
	listener    net.Listener
	connections map[net.Conn]struct{}
	workers     sync.WaitGroup
	stopped     bool
	name        string
	timeout     time.Duration
	listen      func() (net.Listener, error)
	log         *log.Helper
}

func newStreamServer(name string, timeout time.Duration, logger log.Logger, listen func() (net.Listener, error)) *streamServer {
	return &streamServer{name: name, timeout: timeout, listen: listen, connections: make(map[net.Conn]struct{}), log: log.NewHelper(logger)}
}
func (s *streamServer) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil
	}
	lis, err := s.listen()
	if err != nil {
		s.mu.Unlock()
		return err
	}
	s.listener = lis
	s.mu.Unlock()
	s.log.Infof("[%s] listening on %s", s.name, lis.Addr())
	for {
		conn, err := lis.Accept()
		if err != nil {
			s.mu.Lock()
			stopped := s.stopped
			s.mu.Unlock()
			if stopped || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		s.mu.Lock()
		if s.stopped {
			s.mu.Unlock()
			conn.Close()
			return nil
		}
		s.connections[conn] = struct{}{}
		s.workers.Add(1)
		s.mu.Unlock()
		go s.echo(conn)
	}
}
func (s *streamServer) echo(conn net.Conn) {
	defer s.workers.Done()
	defer func() { conn.Close(); s.mu.Lock(); delete(s.connections, conn); s.mu.Unlock() }()
	buf := make([]byte, 4096)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(s.timeout)); err != nil {
			return
		}
		n, err := conn.Read(buf)
		if n > 0 {
			if deadlineErr := conn.SetWriteDeadline(time.Now().Add(s.timeout)); deadlineErr != nil {
				return
			}
			if writeErr := writeAll(conn, buf[:n]); writeErr != nil {
				return
			}
		}
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
				s.log.Debugf("[%s] connection ended: %v", s.name, err)
			}
			return
		}
	}
}
func writeAll(w io.Writer, p []byte) error {
	for len(p) > 0 {
		n, err := w.Write(p)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		p = p[n:]
	}
	return nil
}
func (s *streamServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	s.stopped = true
	if s.listener != nil {
		s.listener.Close()
	}
	for conn := range s.connections {
		conn.Close()
	}
	s.mu.Unlock()
	done := make(chan struct{})
	go func() { s.workers.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
