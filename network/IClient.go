package network

type IClient interface {
	Connect() IConn
	Close(err error)
}
