package tcp

import (
	"context"
	"errors"
	"github.com/yaice-rx/yaice/log"
	"github.com/yaice-rx/yaice/network"
	"github.com/yaice-rx/yaice/utils"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
)

type Server struct {
	sync.Mutex
	type_        network.ServeType
	connCount    int32
	listener     *net.TCPListener
	cancel       context.CancelFunc
	ctx          context.Context
	cones        map[uint64]network.IConn
	receiveQueue chan network.TransitData
	pkg          network.IPacket
}

func NewServer() network.IServer {
	s := &Server{
		type_:        network.TypeServer,
		connCount:    0,
		receiveQueue: make(chan network.TransitData, 1000),
		cones:        make(map[uint64]network.IConn),
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.ctx = ctx
	return s
}

func (s *Server) Listen(packet network.IPacket, startPort int, endPort int) int {
	s.pkg = packet
	port := make(chan int)
	defer close(port)
	for i := startPort; i < endPort; i++ {
		go func() {
			tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(i))
			if nil != err {
				port <- -1
				return
			}
			listener, err := net.ListenTCP("tcp", tcpAddr)
			if nil != err {
				port <- -1
				return
			}
			port <- i
			s.listener = listener
			for {
				tcpConn, err := listener.AcceptTCP()
				if nil != err || nil == tcpConn {
					continue
				}
				atomic.AddInt32(&s.connCount, 1)
				//这一步分拆读取和写入
				conn := NewConn(tcpConn, packet, s.receiveQueue, nil, s.ctx, s.cancel)
				s.cones[conn.GetGuid()] = conn
				go conn.Receive()
				go conn.Write()
			}
		}()
		portData := <-port
		if portData > 0 {
			return portData
		}
	}
	return -1
}

// SendProtobuf
//
//	@Description: 发送协议体(protobuf)
//	@receiver c
//	@param message
//	@return error
func (s *Server) SendProtobuf(sessionGuid uint64, message proto.Message) error {
	conn := s.cones[sessionGuid]
	if conn != nil {
		data, err := proto.Marshal(message)
		protoId := utils.ProtocalNumber(utils.GetProtoName(message))
		if err != nil {
			log.AppLogger.Error("发送消息时，序列化失败 : "+err.Error(), zap.Int32("MessageId", protoId))
			return err
		}
		conn.GetSendChannel() <- s.pkg.Pack(sessionGuid, conn.GetServerAck(), network.TransitData{MsgId: protoId, Data: data})
		return nil
	}
	return errors.New("not found session")
}

// SendByte
//
//	@Description: 发送组装好的协议，但是加密始终是在组装包的时候完成加密功能
//	@receiver c
//	@param message
//	@return error
func (s *Server) SendByte(sessionGuid uint64, message []byte) error {
	conn := s.cones[sessionGuid]
	if conn != nil {
		conn.GetSendChannel() <- message
	}
	return errors.New("not found session")
}

// GetReceiveQueue 获取接收队列
//
//	@Description:
//	@receiver s
//	@return chan
func (s *Server) GetReceiveQueue() chan network.TransitData {
	return s.receiveQueue
}

// Close
//
//	@Description: 网络关闭
//	@receiver s
func (s *Server) Close() {
	for _, conn := range s.cones {
		conn.Close()
	}
}
