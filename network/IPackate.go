package network

type IPacket interface {
	GetHeadLen() uint32
	Pack(data TransitData,isPos int64) []byte
	Unpack(binaryData []byte ) (IMessage, error, func(conn IConn))
}
