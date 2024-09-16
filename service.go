package yaice

import (
	"context"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/tcp"
	"github.com/yaice-rx/yaice/router"
	"google.golang.org/protobuf/proto"
)

// 服务运行状态

type IService interface {
	RegisterProtoHandler(message proto.Message, handler func(conn network.IConn, content []byte))
	Listen(packet network.IPacket, network string, startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) (network.IServer, int)
	Dial(packet network.IPacket, network string, address string, options network.IOptions, reConnCallBackFunc func(conn network.IConn, err error)) network.IClient
	ProtoHandler(msgData network.TransitData)
}

type service struct {
	cancel      context.CancelFunc
	routerMgr   router.IRouter
	serviceType int
}

// NewService
//
//	@Description: 服务初始化
//	@return IService
func NewService() IService {
	return &service{
		routerMgr: router.RouterMgr,
	}
}

// RegisterProtoHandler 注册网络处理方法
//
//	@Description:
//	@receiver s
//	@param message
//	@param handler
func (s *service) RegisterProtoHandler(message proto.Message, handler func(conn network.IConn, content []byte)) {
	s.routerMgr.AddRouter(message, handler)
}

// ProtoHandler
//
//	@Description: 消息处理
//	@receiver s
//	@param msgData
func (s *service) ProtoHandler(msgData network.TransitData) {
	s.routerMgr.ExecRouterFunc(msgData)
}

// Dial
//
//	@Description: 连接网络
//	@receiver s
//	@param packet packet 网络包的协议处理方式，如果传输为nil，则采用默认的方式
//	@param network_ 网络连接方式
//	@param address 地址
//	@param options 传递参数
//	@param callFunc 回调函数
//	@return network.IConn
func (s *service) Dial(packet network.IPacket, network_ string, address string, options network.IOptions, callFunc func(conn network.IConn, err error)) network.IClient {
	if packet == nil {
		packet = tcp.NewPacket()
	}
	switch network_ {
	case "tcp", "tcp4", "tcp6":
		client := tcp.NewClient(packet, address, options, callFunc)
		client.Connect()
		return client
	}
	return nil
}

// Listen
//
//	@Description: 监听网络
//	@receiver s
//	@param packet	网络包的协议处理方式，如果传输为nil，则采用默认的方式
//	@param network_	网络连接方式
//	@param startPort	监听端口范围开始
//	@param endPort	监听端口范围结束
//	@param isAllowConnFunc 是否允许连接
//	@return int
func (s *service) Listen(packet network.IPacket, network_ string, startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) (network.IServer, int) {
	if packet == nil {
		packet = tcp.NewPacket()
	}
	switch network_ {
	case "tcp", "tcp4", "tcp6":
		server := tcp.NewServer()
		return server, server.Listen(packet, startPort, endPort)
	}
	return nil, 0
}
