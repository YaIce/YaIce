package network

type TransitData struct {
	Conn  IConn
	MsgId int32
	Data  []byte
}
