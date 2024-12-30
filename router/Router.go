package router

import (
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
	"google.golang.org/protobuf/proto"
	"sync"
)

type router struct {
	sync.RWMutex
	routers        map[int32]func(conn network.IConn, content []byte)
	msgQueueRouter map[int32]func(content []byte)
}

func NewRouterMgr() IRouter {
	return &router{
		routers: make(map[int32]func(conn network.IConn, content []byte)),
	}
}

func (r *router) AddRouter(msgObj proto.Message, handler func(conn network.IConn, content []byte)) {
	r.Lock()
	defer r.Unlock()
	msgName := utils.GetProtoName(msgObj)
	protocolNum := utils.ProtocalNumber(msgName)
	r.routers[protocolNum] = handler
}

func (r *router) RegisterMQ(msgQueueName string, handler func(content []byte)) {
	r.Lock()
	defer r.Unlock()
	protocolNum := utils.ProtocalNumber(msgQueueName)
	r.msgQueueRouter[protocolNum] = handler
}

func (r *router) ExecRouterFunc(conn network.IConn, message network.TransitData) {
	r.Lock()
	defer r.Unlock()
	handler := r.routers[message.MsgId]
	if handler != nil {
		handler(conn, message.Data)
	}
}
