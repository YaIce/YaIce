package network

type IPacket interface {
	GetHeadLen() uint32
	Pack(playerGuid uint64, isPos uint64, data TransitData) []byte
	Unpack(binaryData []byte) (IMessage, error)
}
