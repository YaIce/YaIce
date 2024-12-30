package kcpNetwork

import (
	"context"
	"fmt"
	"github.com/xtaci/kcp-go/v5"
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"go.uber.org/zap"
	"time"
)

type KCPClient struct {
	type_            network.ServeType
	dialRetriesCount int32
	address          string
	conn             network.IConn
	packet           network.IPacket
	opt              network.IOptions
	ctx              context.Context
	cancel           context.CancelFunc
	callFunc         func(conn network.IConn, err error)
}

func NewClient(packet network.IPacket, address string, opt network.IOptions, callFunc func(conn network.IConn, err error)) network.IClient {
	c := &KCPClient{
		type_:            network.Serve_Client,
		address:          address,
		packet:           packet,
		opt:              opt,
		dialRetriesCount: 0,
		callFunc:         callFunc,
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel
	return c
}

func (c *KCPClient) Connect() network.IConn {
LOOP:
	tcpConn, err := kcp.DialWithOptions(c.address, nil, 0, 0)
	if err != nil {
		time.Sleep(3 * time.Second)
		if c.opt.GetMaxRetires() < c.dialRetriesCount {
			log.AppLogger.Error("网络重连失败:"+err.Error(), zap.String("function", "network.tcp.Client.Connect"))
			return nil
		}
		log.AppLogger.Warn(fmt.Sprintf("第{%d}网络重连中", c.dialRetriesCount))
		c.dialRetriesCount += 1
		goto LOOP
	}
	//连接上的时候，重置连接次数
	c.dialRetriesCount = 0
	c.conn = NewConn(c, tcpConn, c.packet, c.opt, network.Serve_Client, c.ctx, c.cancel)
	//读取网络通道数据
	go c.conn.Start()
	return c.conn
}

func (c *KCPClient) ReConnect() network.IConn {
	return c.Connect()
}

func (c *KCPClient) Close(err error) {
	c.cancel()
	//设置当前客户端的状态
	c.callFunc(c.conn, err)
	c.conn.Close()
}
