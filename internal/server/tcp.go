package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/conf"
	"net"
)

type TCPServer struct{ *streamServer }

func NewTCPServer(c *conf.Server_TCP, logger log.Logger) *TCPServer {
	return &TCPServer{newStreamServer("TCP", conf.Timeout(c.Timeout), logger, func() (net.Listener, error) { return net.Listen("tcp", c.Addr) })}
}
