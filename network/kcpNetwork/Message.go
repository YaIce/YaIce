package kcpNetwork

type Message struct {
	ID   int32
	Data []byte
}

type MsgQueueData struct {
	Queue chan Message
}

//获取消息ID
func (this *Message) GetMsgId() int32 {
	return this.ID
}

//获取消息内容
func (this *Message) GetData() []byte {
	return this.Data
}

//获取消息内容
func (this *Message) GetIsPos() int64 {
	return 0
}
