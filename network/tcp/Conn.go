package tcp

import (
	"context"
	"errors"
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/router"
	"github.com/yaice-rx/yaice/utils"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"sync/atomic"
	"time"
)

type Conn struct {
	type_        network.ServeType
	guid         uint64
	isClosed     bool
	times        int64
	pkg          network.IPacket
	conn         *net.TCPConn
	serve        interface{}
	data         interface{}
	isPos        int64
	sendQueue    chan []byte
	receiveQueue chan network.TransitData
	router       router.IRouter
	opt          network.IOptions
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewConn(serve interface{}, conn *net.TCPConn, pkg network.IPacket, opt network.IOptions, type_ network.ServeType,
	ctx context.Context, cancelFunc context.CancelFunc) network.IConn {
	conn_ := &Conn{
		type_:        type_,
		serve:        serve,
		guid:         utils.GenSonyflakeToo(),
		conn:         conn,
		pkg:          pkg,
		router:       router.NewRouterMgr(),
		receiveQueue: make(chan network.TransitData, 5000),
		sendQueue:    make(chan []byte, 5000),
		times:        time.Now().Unix(),
		isClosed:     false,
		opt:          opt,
		ctx:          ctx,
		cancel:       cancelFunc,
	}
	//数据写入
	go func(conn_ *Conn) {
		for data := range conn_.sendQueue {
		LOOP:
			_, err := conn_.conn.Write(data)
			//判断客户端，如果不是主动关闭，而是网络抖动的时候 多次连接
			if conn_.type_ == network.Serve_Client && nil != err {
				if conn_.serve.(*TCPClient).dialRetriesCount > conn_.serve.(*TCPClient).opt.GetMaxRetires() && err != nil {
					conn_.serve.(*TCPClient).Close(err)
				}
				if conn_.serve.(*TCPClient).dialRetriesCount <= conn_.serve.(*TCPClient).opt.GetMaxRetires() && err != nil {
					conn_.serve.(*TCPClient).dialRetriesCount += 1
					goto LOOP
				}
			}
			//发送成功
			if conn_.type_ == network.Serve_Client {
				conn_.serve.(*TCPClient).dialRetriesCount = 0
			}
		}
	}(conn_)
	//数据读取
	go func(conn_ *Conn) {
		for data := range conn_.receiveQueue {
			if data.MsgId != 0 {
				conn_.router.ExecRouterFunc(conn_, data)
			}
		}
	}(conn_)
	return conn_
}

func (c *Conn) Close() {
	c.isClosed = true
	//如果当前链接已经关闭
	if c.type_ == network.Serve_Server {
		//断开连接减少对应的连接
		atomic.AddInt32(&c.serve.(*Server).connCount, int32(-1))
	}
	//关闭接收通道
	close(c.receiveQueue)
	//关闭发送通道
	close(c.sendQueue)
}

func (c *Conn) GetGuid() uint64 {
	return c.guid
}

func (c *Conn) GetConn() interface{} {
	return c.conn
}

func (c *Conn) GetIsPos() int64 {
	return c.isPos
}

func (c *Conn) getRouter() router.IRouter {
	return c.router
}

// 发送协议体
func (c *Conn) Send(message proto.Message) error {
	select {
	case <-c.ctx.Done():
		c.isClosed = true
		return errors.New("send msg(proto) channel It has been closed ... ")
	default:
		if c.isClosed != true {
			data, err := proto.Marshal(message)
			protoId := utils.ProtocalNumber(utils.GetProtoName(message))
			if err != nil {
				log.AppLogger.Error("发送消息时，序列化失败 : "+err.Error(), zap.Int32("MessageId", protoId))
				return err
			}
			c.sendQueue <- c.pkg.Pack(network.TransitData{protoId, data}, c.isPos)
		}
		return nil
	}
}

// 发送组装好的协议，但是加密始终是在组装包的时候完成加密功能
func (c *Conn) SendByte(message []byte) error {
	select {
	case <-c.ctx.Done():
		c.isClosed = true
		return errors.New("send msg(proto) channel It has been closed ... ")
	default:
		if c.isClosed != true {
			c.sendQueue <- message
		}
		return nil
	}
}

func (c *Conn) Start() {
	for {
		select {
		case <-c.ctx.Done():
			log.AppLogger.Info("conn It has been closed ... ")
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
					if c.type_ == network.Serve_Client {
						c.serve.(*TCPClient).Close(err)
					}
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
				msgData, err, func_ := c.pkg.Unpack(contentData)
				if msgData == nil {
					if err != nil {
						log.AppLogger.Info("network io read data err - 1:" + err.Error())
					}
					continue
				}
				if func_ != nil {
					func_(c)
				}
				c.isPos = msgData.GetIsPos()
				//写入通道数据
				c.receiveQueue <- network.TransitData{
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

func (c *Conn) GetCreateTime() int64 {
	return c.times
}

func (c *Conn) GetOptions() network.IOptions {
	return c.opt
}
