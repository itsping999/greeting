package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/wire"
	"github.com/itsping999/greeting/internal/conf"
	"github.com/itsping999/greeting/internal/service"
)

var ProviderSet = wire.NewSet(NewServers)

type Servers struct{ items []transport.Server }

func (s *Servers) All() []transport.Server { return s.items }

func NewServers(c *conf.Server, logger log.Logger, svc *service.GreeterService) (*Servers, error) {
	if err := conf.Validate(c); err != nil {
		return nil, err
	}
	s := &Servers{}
	if v := c.Http; v != nil && v.Enabled {
		s.items = append(s.items, NewHTTPServer(v, logger, svc))
	}
	if v := c.Grpc; v != nil && v.Enabled {
		s.items = append(s.items, NewGRPCServer(v, logger, svc))
	}
	if v := c.Ws; v != nil && v.Enabled {
		s.items = append(s.items, NewWebsocketServer(v, logger, svc))
	}
	if v := c.Mqtt; v != nil && v.Enabled {
		s.items = append(s.items, NewMQTTServer(v, logger, svc))
	}
	if v := c.Rabbitmq; v != nil && v.Enabled {
		s.items = append(s.items, NewRabbitMQServer(v, logger, svc))
	}
	if v := c.Tcp; v != nil && v.Enabled {
		s.items = append(s.items, NewTCPServer(v, logger))
	}
	if v := c.Udp; v != nil && v.Enabled {
		s.items = append(s.items, NewUDPServer(v, logger))
	}
	if v := c.Kcp; v != nil && v.Enabled {
		s.items = append(s.items, NewKCPServer(v, logger))
	}
	return s, nil
}
