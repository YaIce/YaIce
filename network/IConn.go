package network

import (
	"github.com/yaice-rx/yaice/router"
	"google.golang.org/protobuf/proto"
)

type IConn interface {
	GetGuid() uint64
	Close()
	Start()
	Send(message proto.Message) error
	SendByte(message []byte) error
	GetConn() interface{}
	getRouter() router.IRouter
	GetIsPos() int64
	GetCreateTime() int64
	GetOptions() IOptions
}
