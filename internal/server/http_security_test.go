package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/service"
)

func TestHTTPDoesNotExposeDefaultServeMux(t *testing.T) {
	originalMux := http.DefaultServeMux
	globalMux := http.NewServeMux()
	http.DefaultServeMux = globalMux
	t.Cleanup(func() { http.DefaultServeMux = originalMux })

	globalCalls := 0
	globalHandler := func(w http.ResponseWriter, _ *http.Request) {
		globalCalls++
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "global diagnostics must not be exposed")
	}
	globalMux.HandleFunc("/debug/pprof/", globalHandler)
	globalMux.HandleFunc("/healthz", globalHandler)

	logger := log.NewStdLogger(io.Discard)
	system := service.NewSystemService(nil, nil, nil)
	srv := newHTTPServer(&HTTPConfig{Enabled: true}, nil, nil, logger, system)
	for _, test := range []struct {
		name   string
		method string
		path   string
		status int
	}{
		{"unregistered global route", http.MethodGet, "/debug/pprof/", http.StatusNotFound},
		{"wrong method on registered route", http.MethodPost, "/healthz", http.StatusMethodNotAllowed},
		{"registered route remains available", http.MethodGet, "/healthz", http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			srv.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
			if response.Code != test.status {
				t.Errorf("status = %d, want %d", response.Code, test.status)
			}
		})
	}
	if globalCalls != 0 {
		t.Fatalf("HTTP server reached global handlers %d times", globalCalls)
	}
}
