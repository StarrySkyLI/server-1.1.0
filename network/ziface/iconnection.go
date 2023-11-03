package ziface

import "net"

// 定义链接模块的抽象层
type IConnection interface {
	Start()
	Stop()
	//获取当前链接的绑定socket conn
	GetTcpConnection() *net.TCPConn
	//获取当前链接模块的链接ID
	GetConnId() uint32
	//获取远程客户端的 tcp状态 Ip port
	RemoteAddr() net.Addr
	//发送数据，将数据发送给远程客户端
	SendMsg(msgId uint32, data []byte) error
	//设置链接属性
	Setproperty(key string, value interface{})
	//获取链接属性
	Getproperty(key string) (interface{}, error)
	//移除链接属性
	Removeproperty(key string)
}

// 定义一个处理链接业务的方法
type HandleFunc func(*net.TCPConn, []byte, int) error
