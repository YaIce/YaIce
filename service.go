package yaice

import (
	"context"
	"github.com/yaice-rx/yaice/config"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/network/kcpNetwork"
	"github.com/yaice-rx/yaice/network/tcp"
)

// 服务运行状态
var shutdown = make(chan bool, 1)

type IService interface {
	Listen(packet network.IPacket, network string, startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) int
	Dial(packet network.IPacket, network string, address string, options network.IOptions, reConnCallBackFunc func(conn network.IConn, err error)) network.IConn
	Close()
}

type service struct {
	cancel      context.CancelFunc
	configMgr   config.IConfig
	ServiceType int
}

/**
 * @param endpoints 集群管理中心连接节点
 */
func NewService() IService {
	return &service{
		configMgr: config.ConfInstance(),
	}
}

/**
 * 连接网络
 * @param network.IPacket  packet 网络包的协议处理方式，如果传输为nil，则采用默认的方式
 * @param network string 网络连接方式
 * @param address string 地址
 * @param options 最大连接次数
 */
func (s *service) Dial(packet network.IPacket, network_ string, address string, options network.IOptions, callFunc func(conn network.IConn, err error)) network.IConn {
	if packet == nil {
		packet = tcp.NewPacket()
	}
	switch network_ {
	case "kcpNetwork":
		return kcpNetwork.NewClient(packet, address, options, callFunc).Connect()
	case "tcp", "tcp4", "tcp6":
		return tcp.NewClient(packet, address, options, callFunc).Connect()
	}
	return nil
}

/**
 * @param network.IPacket  packet 网络包的协议处理方式，如果传输为nil，则采用默认的方式
 * @param string network 网络连接方式
 * @param int startPort 监听端口范围开始
 * @param int endPort 监听端口范围结束
 * @param func isAllowConnFunc  限制连接数，超过连接数的时候，由上层逻辑通知，底层不予维护
 */
func (s *service) Listen(packet network.IPacket, network_ string, startPort int, endPort int, isAllowConnFunc func(conn interface{}) bool) int {
	if packet == nil {
		packet = tcp.NewPacket()
	}
	switch network_ {
	case "kcpNetwork":
		serverMgr := kcpNetwork.NewServer()
		return serverMgr.Listen(packet, startPort, endPort, isAllowConnFunc)
	case "tcp", "tcp4", "tcp6":
		serverMgr := tcp.NewServer()
		return serverMgr.Listen(packet, startPort, endPort, isAllowConnFunc)
	}
	return 0
}

/**
 * 关闭集群服务
 */
func (s *service) Close() {

}
