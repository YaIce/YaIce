package network

type IServer interface {
	Listen(packet IPacket, startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) int
	Close()
}
