package conf

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

func Timeout(value *durationpb.Duration) time.Duration {
	if value == nil {
		return 5 * time.Second
	}
	return value.AsDuration()
}

// Validate checks enabled transports before acquiring any resources.
func Validate(c *Server) error {
	if c == nil {
		return fmt.Errorf("server configuration is required")
	}
	count := 0
	check := func(name, addr string, timeout *durationpb.Duration) error {
		count++
		if _, _, err := net.SplitHostPort(addr); err != nil {
			return fmt.Errorf("%s.addr must be host:port", name)
		}
		if timeout != nil {
			if err := timeout.CheckValid(); err != nil || timeout.AsDuration() <= 0 {
				return fmt.Errorf("%s.timeout must be positive", name)
			}
		}
		return nil
	}
	if v := c.Http; v != nil && v.Enabled {
		if err := check("http", v.Addr, v.Timeout); err != nil {
			return err
		}
	}
	if v := c.Grpc; v != nil && v.Enabled {
		if err := check("grpc", v.Addr, v.Timeout); err != nil {
			return err
		}
	}
	if v := c.Ws; v != nil && v.Enabled {
		if err := check("ws", v.Addr, v.Timeout); err != nil {
			return err
		}
		if !strings.HasPrefix(v.Path, "/") {
			return fmt.Errorf("ws.path must start with /")
		}
	}
	if v := c.Tcp; v != nil && v.Enabled {
		if err := check("tcp", v.Addr, v.Timeout); err != nil {
			return err
		}
	}
	if v := c.Udp; v != nil && v.Enabled {
		if err := check("udp", v.Addr, v.Timeout); err != nil {
			return err
		}
	}
	if v := c.Kcp; v != nil && v.Enabled {
		if err := check("kcp", v.Addr, v.Timeout); err != nil {
			return err
		}
		if v.Key == "" || v.Salt == "" {
			return fmt.Errorf("KCP requires a key and salt; set GREETING_KCP_KEY and GREETING_KCP_SALT")
		}
	}
	for _, v := range []struct {
		name, addr string
		enabled    bool
	}{
		{"mqtt", c.GetMqtt().GetAddr(), c.GetMqtt().GetEnabled()},
		{"rabbitmq", c.GetRabbitmq().GetAddr(), c.GetRabbitmq().GetEnabled()},
	} {
		if !v.enabled {
			continue
		}
		count++
		u, err := url.Parse(v.addr)
		if err != nil || u.Host == "" {
			return fmt.Errorf("%s.addr must be a broker URL", v.name)
		}
		if v.name == "mqtt" && u.Scheme != "tcp" && u.Scheme != "ssl" {
			return fmt.Errorf("mqtt.addr must use tcp or ssl")
		}
		if v.name == "rabbitmq" && u.Scheme != "amqp" && u.Scheme != "amqps" {
			return fmt.Errorf("rabbitmq.addr must use amqp or amqps")
		}
	}
	if count == 0 {
		return fmt.Errorf("enable at least one server transport")
	}
	return nil
}
