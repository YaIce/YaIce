package tcp

import (
	"context"
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
	"io"
	"log/slog"
	"net"
	"time"
)

type Conn struct {
	guid         uint64
	isClosed     bool
	times        int64
	pkg          network.IPacket
	conn         *net.TCPConn
	serve        interface{}
	data         interface{}
	ServerAck    uint64
	ClientAck    uint64
	receiveQueue chan network.TransitData
	sendQueue    chan []byte
	opt          network.IOptions
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewConn(conn *net.TCPConn, pkg network.IPacket, receiveQueue chan network.TransitData, opt network.IOptions,
	ctx context.Context, cancelFunc context.CancelFunc) network.IConn {
	return &Conn{
		guid:         utils.GenSonyflakeToo(),
		conn:         conn,
		pkg:          pkg,
		times:        time.Now().Unix(),
		isClosed:     false,
		receiveQueue: receiveQueue,
		sendQueue:    make(chan []byte, 1000),
		opt:          opt,
		ctx:          ctx,
		cancel:       cancelFunc,
	}
}

func (c *Conn) Close() {
	c.isClosed = true
}

func (c *Conn) GetGuid() uint64 {
	return c.guid
}

func (c *Conn) GetConn() interface{} {
	return c.conn
}

func (c *Conn) Receive() {
	for {
		select {
		case <-c.ctx.Done():
			slog.Info("conn It has been closed ... ")
			return
		default:
			//读取限时
			if err := c.conn.SetReadDeadline(time.Now().Add(time.Minute * 70)); err != nil {
				return
			}
			//1 先读出流中的head部分
			headData := make([]byte, c.pkg.GetHeadLen())
			_, err := io.ReadAtLeast(c.conn, headData[:4], 4) //io.ReadFull(c.conn, headData) //ReadFull 会把msg填充满为止
			if err != nil {
				if err != io.EOF {
					return
				}
				break
			}
			msgLen := utils.BytesToInt(headData)
			if msgLen > 0 {
				//msg 是有data数据的，需要再次读取data数据
				contentData := make([]byte, msgLen)
				//根据dataLen从io中读取字节流
				_, err := io.ReadFull(c.conn, contentData)
				if err != nil {
					log.AppLogger.Info("network io read data err - 0:" + err.Error())
					break
				}
				//解压网络数据包
				msgData, err := c.pkg.Unpack(contentData)
				if msgData == nil {
					if err != nil {
						log.AppLogger.Info("network io read data err - 1:" + err.Error())
					}
					continue
				}
				c.ClientAck = msgData.GetIsPos()
				//写入通道数据
				c.receiveQueue <- network.TransitData{
					Conn:  c,
					MsgId: msgData.GetMsgId(),
					Data:  msgData.GetData(),
				}
			}
			if err := c.conn.SetReadDeadline(time.Time{}); err != nil {
				return
			}
		}
	}
}

func (c *Conn) GetServerAck() uint64 {
	return c.ServerAck
}

func (c *Conn) GetClientAck() uint64 {
	return c.ClientAck
}

func (c *Conn) GetSendChannel() chan []byte {
	return c.sendQueue
}

func (c *Conn) Write() {
	for data := range c.sendQueue {
		_, err := c.conn.Write(data)
		if err != nil {
			return
		}
	}
}
