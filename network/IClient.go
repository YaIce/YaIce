package network

import "google.golang.org/protobuf/proto"

type IClient interface {
	Connect() IConn
	Close()
	GetReceiveQueue() chan TransitData
	SendByte(sessionGuid uint64, message []byte) error
	SendProtobuf(sessionGuid uint64, message proto.Message) error
}
