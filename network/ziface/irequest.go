package ziface

//ireqeust接口：实际上是把客户端请求的链接消息和请求数据包装到一个Request中

type IRequest interface {
	//得到当前的链接
	GetConnection() IConnection
	//得到请求的消息数据
	GetData() []byte
	//得到请求的消息数据ID
	GetMsgID() uint32
}
