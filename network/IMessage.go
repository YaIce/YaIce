package network

type IMessage interface {
	GetMsgId() int32       //获取消息ID
	GetPlayerGuid() uint64 //获取玩家的guid
	GetData() []byte       //获取消息内容
	GetIsPos() uint64      //获取当前消息位置
}
