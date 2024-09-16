package tcp

import (
	"context"
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"net"
	"time"
)

type Client struct {
	type_            network.ServeType //网络类型
	dialRetriesCount int32             //拨号重试次数
	address          string            //地址
	conn             network.IConn
	packet           network.IPacket
	opt              network.IOptions
	ctx              context.Context
	cancel           context.CancelFunc
	callFunc         func(conn network.IConn, err error)
	receiveQueue     chan network.TransitData
	sendQueue        chan []byte
}

func NewClient(packet network.IPacket, address string, opt network.IOptions, callFunc func(conn network.IConn, err error)) network.IClient {
	c := &Client{
		address:          address,
		packet:           packet,
		opt:              opt,
		dialRetriesCount: 0,
		callFunc:         callFunc,
		receiveQueue:     make(chan network.TransitData, 1000),
		sendQueue:        make(chan []byte, 1000),
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel
	return c
}

func (c *Client) Connect() network.IConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", c.address)
	if err != nil {
		c.callFunc(c.conn, err)
		return nil
	}
LOOP:
	tcpConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		time.Sleep(3 * time.Second)
		if c.opt.GetMaxRetires() < c.dialRetriesCount {
			c.callFunc(c.conn, err)
			return nil
		}
		c.dialRetriesCount += 1
		goto LOOP
	}
	//连接上的时候，重置连接次数
	c.dialRetriesCount = 0
	c.conn = NewConn(tcpConn, c.packet, c.receiveQueue, c.opt, c.ctx, c.cancel)
	//读取网络通道数据
	go c.conn.Receive()
	//写入通道数据
	go c.conn.Write()
	return c.conn
}

// SendProtobuf
//
//	@Description: 发送协议体(protobuf)
//	@receiver c
//	@param message
//	@return error
func (c *Client) SendProtobuf(sessionGuid uint64, message proto.Message) error {
	data, err := proto.Marshal(message)
	protoId := utils.ProtocalNumber(utils.GetProtoName(message))
	if err != nil {
		log.AppLogger.Error("发送消息时，序列化失败 : "+err.Error(), zap.Int32("MessageId", protoId))
		return err
	}
	c.sendQueue <- c.packet.Pack(sessionGuid, c.conn.GetServerAck(), network.TransitData{MsgId: protoId, Data: data})
	return nil
}

// SendByte
//
//	@Description: 发送组装好的协议，但是加密始终是在组装包的时候完成加密功能
//	@receiver c
//	@param message
//	@return error
func (c *Client) SendByte(sessionGuid uint64, message []byte) error {
	c.sendQueue <- message
	return nil
}

// GetReceiveQueue
//
//	@Description: 获取接收队列
//	@receiver c
//	@return chan
func (c *Client) GetReceiveQueue() chan network.TransitData {
	return c.receiveQueue
}

// ReConnect
//
//	@Description: 重连
//	@receiver c
//	@return network.IConn
func (c *Client) ReConnect() network.IConn {
	return c.Connect()
}

// Close
//
//	@Description: 网络关闭
//	@receiver c
//	@param err
func (c *Client) Close() {
	c.cancel()
	//设置当前客户端的状态
	c.conn.Close()
}
