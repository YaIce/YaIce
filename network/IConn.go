package network

type IConn interface {
	GetGuid() uint64
	Close()
	Receive()
	Write()
	GetConn() interface{}
	GetServerAck() uint64
	GetClientAck() uint64
	GetSendChannel() chan []byte
}
