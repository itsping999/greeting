package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"net"
)

type TCPServer struct{ *streamServer }

func NewTCPServer(c *TCPConfig, logger log.Logger) *TCPServer {
	if c == nil || !c.Enabled {
		return nil
	}

	return &TCPServer{newStreamServer("TCP", Timeout(c.Timeout), logger, func() (net.Listener, error) { return net.Listen("tcp", c.Addr) })}
}
