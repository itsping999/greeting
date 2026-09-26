package conf

import (
	"github.com/itsping999/greeting/internal/data"
	"github.com/itsping999/greeting/internal/pkg/logx"
	"github.com/itsping999/greeting/internal/pkg/tracingx"
	"github.com/itsping999/greeting/internal/server"
)

type Bootstrap struct {
	Server  server.Config   `json:"server" yaml:"server"`
	Data    data.Config     `json:"data" yaml:"data"`
	Logger  logx.Options    `json:"logger" yaml:"logger"`
	Tracing tracingx.Config `json:"tracing" yaml:"tracing"`
}
