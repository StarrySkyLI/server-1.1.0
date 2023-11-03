package znet

import "server-1.1.0/network/ziface"

type Request struct {
	//已经和客户端建立好的链接
	conn ziface.IConnection

	//得到客户端请求的数据
	msg ziface.IMessage
}

func (r *Request) GetConnection() ziface.IConnection {
	return r.conn
}
func (r *Request) GetData() []byte {
	return r.msg.GetData()
}
func (r *Request) GetMsgID() uint32 {
	return r.msg.GetMsgId()
}
