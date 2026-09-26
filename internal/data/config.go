package data

import "github.com/itsping999/greeting/internal/pkg/dbx"
import "github.com/itsping999/greeting/internal/pkg/discoveryx"
import "github.com/itsping999/greeting/internal/pkg/grpcx"

type Config struct {
	Discovery  discoveryx.Config  `json:"discovery" yaml:"discovery"`
	Postgres   dbx.PostgresConfig `json:"postgres" yaml:"postgres"`
	Redis      dbx.RedisConfig    `json:"redis" yaml:"redis"`
	MongoDB    dbx.MongoConfig    `json:"mongodb" yaml:"mongodb"`
	RemoteGRPC grpcx.Config       `json:"remote_grpc" yaml:"remote_grpc"`
}
