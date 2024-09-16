package network

import "google.golang.org/protobuf/proto"

type IServer interface {
	Listen(packet IPacket, startPort int, endPort int) int
	Close()
	SendByte(sessionGuid uint64, message []byte) error
	SendProtobuf(sessionGuid uint64, message proto.Message) error
	GetReceiveQueue() chan TransitData
}
