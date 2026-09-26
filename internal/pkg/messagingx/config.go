package messagingx

type MQTTConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Addr    string `json:"addr" yaml:"addr"`
	Topic   string `json:"topic" yaml:"topic"`
}

type RabbitMQConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Addr       string `json:"addr" yaml:"addr"`
	Exchange   string `json:"exchange" yaml:"exchange"`
	RoutingKey string `json:"routing_key" yaml:"routing_key"`
	Queue      string `json:"queue" yaml:"queue"`
}
