package server

import (
	"crypto/sha1"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/conf"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
	"net"
)

type KCPServer struct{ *streamServer }

func NewKCPServer(c *conf.Server_KCP, logger log.Logger) *KCPServer {
	return &KCPServer{newStreamServer("KCP", conf.Timeout(c.Timeout), logger, func() (net.Listener, error) {
		key := pbkdf2.Key([]byte(c.Key), []byte(c.Salt), 4096, 32, sha1.New)
		block, err := kcp.NewAESBlockCrypt(key)
		if err != nil {
			return nil, err
		}
		return kcp.ListenWithOptions(c.Addr, block, 10, 3)
	})}
}
