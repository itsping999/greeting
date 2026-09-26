package server

import (
	"crypto/sha1"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
	"net"
)

type KCPServer struct{ *streamServer }

func NewKCPServer(c *KCPConfig, logger log.Logger) *KCPServer {
	if c == nil || !c.Enabled {
		return nil
	}

	return &KCPServer{newStreamServer("KCP", Timeout(c.Timeout), logger, func() (net.Listener, error) {
		key := pbkdf2.Key([]byte(c.Key), []byte(c.Salt), 4096, 32, sha1.New)
		block, err := kcp.NewAESBlockCrypt(key)
		if err != nil {
			return nil, err
		}
		return kcp.ListenWithOptions(c.Addr, block, 10, 3)
	})}
}
