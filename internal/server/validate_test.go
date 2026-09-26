package server

import (
	"testing"
	"time"
)

func TestValidateEnabledTransports(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{"missing", nil, false}, {"none", &Config{}, false},
		{"minimal HTTP", &Config{HTTP: HTTPConfig{Enabled: true, Addr: "127.0.0.1:0"}}, true},
		{"disabled brokers ignored", &Config{HTTP: HTTPConfig{Enabled: true, Addr: "127.0.0.1:0"}, MQTT: MQTTConfig{Addr: "invalid"}}, true},
		{"bad address", &Config{HTTP: HTTPConfig{Enabled: true, Addr: "invalid"}}, false},
		{"negative timeout", &Config{HTTP: HTTPConfig{Enabled: true, Addr: "127.0.0.1:0", Timeout: -time.Second}}, false},
		{"missing KCP key", &Config{KCP: KCPConfig{Enabled: true, Addr: "127.0.0.1:0"}}, false},
		{"bad broker URL", &Config{MQTT: MQTTConfig{Enabled: true, Addr: "http://localhost"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if (err == nil) != tt.valid {
				t.Fatalf("valid=%v, error=%v", tt.valid, err)
			}
		})
	}
}
