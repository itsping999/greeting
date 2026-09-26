package server

import "time"

type Config struct {
	TCP        TCPConfig        `json:"tcp" yaml:"tcp"`
	UDP        UDPConfig        `json:"udp" yaml:"udp"`
	KCP        KCPConfig        `json:"kcp" yaml:"kcp"`
	HTTP       HTTPConfig       `json:"http" yaml:"http"`
	GRPC       GRPCConfig       `json:"grpc" yaml:"grpc"`
	Middleware MiddlewareConfig `json:"middleware" yaml:"middleware"`
	WebSocket  WebSocketConfig  `json:"websocket" yaml:"websocket"`
	MQTT       MQTTConfig       `json:"mqtt" yaml:"mqtt"`
	RabbitMQ   RabbitMQConfig   `json:"rabbitmq" yaml:"rabbitmq"`
}

type MiddlewareConfig struct {
	RateLimitEnabled bool `json:"ratelimit_enabled" yaml:"ratelimit_enabled"`
}

// StreamConfig is shared by byte-echo transports.
type StreamConfig struct {
	Enabled bool          `json:"enabled" yaml:"enabled"`
	Addr    string        `json:"addr" yaml:"addr"`
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}
type TCPConfig StreamConfig
type UDPConfig StreamConfig
type KCPConfig struct {
	Enabled bool          `json:"enabled" yaml:"enabled"`
	Addr    string        `json:"addr" yaml:"addr"`
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	Key     string        `json:"key" yaml:"key"`
	Salt    string        `json:"salt" yaml:"salt"`
}
