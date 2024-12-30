package router

import (
	"github.com/yaice-rx/yaice/network"
	"google.golang.org/protobuf/proto"
)

type IRouter interface {
	AddRouter(msgObj proto.Message, handler func(conn network.IConn, content []byte))
	RegisterMQ(msgQueueName string, handler func(content []byte))
	ExecRouterFunc(conn network.IConn, message network.TransitData)
}
