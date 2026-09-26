package data

import (
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/itsping999/greeting/internal/biz"
)

var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewGreeterEventPublisher)

// Data owns process-local example storage. Restarting the service clears it.
type Data struct {
	mu       sync.RWMutex
	nextID   int64
	greeters map[int64]biz.Greeter
}

func NewData(logger log.Logger) *Data {
	log.NewHelper(logger).Info("using in-memory example storage; records are lost on restart")
	return &Data{greeters: make(map[int64]biz.Greeter)}
}
