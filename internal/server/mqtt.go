package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/itsping999/greeting/internal/conf"
	"github.com/itsping999/greeting/internal/protocol"
	"github.com/itsping999/greeting/internal/service"
	"github.com/tx7do/kratos-transport/broker"
	"github.com/tx7do/kratos-transport/broker/mqtt"
	"time"
)

func NewMQTTServer(c *conf.Server_MQTT, logger log.Logger, svc *service.GreeterService) *BrokerServer {
	return &BrokerServer{broker: mqtt.NewBroker(broker.WithAddress(c.Addr), broker.WithCodec("json"), mqtt.WithConnectTimeout(5*time.Second), mqtt.WithWriteTimeout(5*time.Second)), name: "MQTT", topic: protocol.MQTTRequestTopic, replyPrefix: protocol.MQTTReplyPrefix, handle: svc.OnMQTTMessage, log: log.NewHelper(logger), opts: []broker.SubscribeOption{mqtt.WithSubscribeQos(1)}}
}
