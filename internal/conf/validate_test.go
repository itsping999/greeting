package conf

import (
	"google.golang.org/protobuf/types/known/durationpb"
	"testing"
	"time"
)

func TestValidateEnabledTransports(t *testing.T) {
	tests := []struct {
		name   string
		config *Server
		valid  bool
	}{
		{"missing", nil, false}, {"none", &Server{}, false},
		{"minimal HTTP", &Server{Http: &Server_HTTP{Enabled: true, Addr: "127.0.0.1:0"}}, true},
		{"disabled brokers ignored", &Server{Http: &Server_HTTP{Enabled: true, Addr: "127.0.0.1:0"}, Mqtt: &Server_MQTT{Addr: "invalid"}}, true},
		{"bad address", &Server{Http: &Server_HTTP{Enabled: true, Addr: "invalid"}}, false},
		{"negative timeout", &Server{Http: &Server_HTTP{Enabled: true, Addr: "127.0.0.1:0", Timeout: durationpb.New(-time.Second)}}, false},
		{"missing KCP key", &Server{Kcp: &Server_KCP{Enabled: true, Addr: "127.0.0.1:0"}}, false},
		{"bad broker URL", &Server{Mqtt: &Server_MQTT{Enabled: true, Addr: "http://localhost"}}, false},
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
