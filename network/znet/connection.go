package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"server-1.1.0/network/utils"
	"server-1.1.0/network/ziface"
	"sync"
)

// 链接模块
type Connection struct {
	//当前Conn隶属哪个Server
	TcpServer ziface.IServer
	//当前链接的socket TCP套接字
	Conn *net.TCPConn
	//链接的ID
	ConnID uint32
	//当前链接的状态
	isClosed bool
	//当前绑定的处理业务方法API
	handleAPI ziface.HandleFunc
	//告知当前链接已经退出、停止 channel(chan 类型 里面是bool值） 由Reader告知
	ExitChan chan bool
	//无缓冲管道，用于读写goroutine之间的信息通信
	msgChan chan []byte
	//消息的管理MsgID和对应的处理业务API关系
	Msghandler ziface.IMsgHanle
	//链接属性集合
	property map[string]interface{}
	//保护链接属性的锁
	propertyLock sync.RWMutex
}

// 初始化链接模块的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHanle) *Connection {
	c := &Connection{
		TcpServer:  server,
		Conn:       conn,
		ConnID:     connID,
		Msghandler: msgHandler,
		isClosed:   false,
		msgChan:    make(chan []byte),
		ExitChan:   make(chan bool, 1),
		property:   make(map[string]interface{}),
	}
	//将conn加入connManager中
	c.TcpServer.GetConnMgr().Add(c)
	return c

}

// 业务读数据的方法
func (c *Connection) StartReader() {
	fmt.Println("[reader goroutine is running..]")
	defer fmt.Println("[Reader is exit],connID= ", c.ConnID, "remote addr is", c.RemoteAddr().String())
	defer c.Stop()
	for {
		////读取客户端的数据到buf中，最大配置文件获取
		//buf := make([]byte, utils.GlobalObject.MaxPacketSize)
		//_, err := c.Conn.Read(buf)
		//if err != nil {
		//	fmt.Println("recv buf err", err)
		//	//这里不用return 不用break 用continue 原因在于还要接着执行下面的方法
		//	continue
		//}
		//创建一个拆包解包的对象
		dp := NewDataPack()
		//读取客户端的Msg Head二进制流8字节
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTcpConnection(), headData); err != nil {
			fmt.Println("read msg head error", err)
			break
		}

		//拆包得msgid和msgDatalen放在msg消息中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error", err)
			break
		}

		//根据datalen再次读data放在msg.data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTcpConnection(), data); err != nil {
				fmt.Println("read msg data error", err)
				break
			}
		}
		msg.SetData(data)

		//得到当前conn数据的Request请求数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经开启了工作池机制，将消息送给Worker工作池处理即可
			c.Msghandler.SendMsgToTaskQueue(&req)
		} else {
			//从路由中，找到注册绑定的Conn对应的Router调用
			//根据绑定好的MsgID找到对应处理api业务执行

			go c.Msghandler.DoMsgHandler(&req)

		}

	}
}

// 写消息的goroutine，专门发送给客户端消息的模块
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println("[conn Writer exit]", c.RemoteAddr().String())
	//不断的阻塞的等待channel的消息，进行写给客户端
	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data err", err)
				return
			}
		case <-c.ExitChan:
			//代表Reader已经退出，此时Writer也要退出
			return
		}
	}
}

func (c *Connection) Start() {
	fmt.Println("Conn Start()...ConnID=", c.ConnID)
	//启动从当前链接的读数据业务
	go c.StartReader()
	// 启动从当前链接写数据的业务
	go c.StartWriter()
	//按照开发者传递进来的 创建链接之后需要调用的处理业务，执行对应Hook函数
	c.TcpServer.CallOnConnStart(c)

}
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()...ConnID=", c.ConnID)

	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	//调用开发者注册进来的 销毁链接之前需要调用的处理业务，执行对应Hook函数
	c.TcpServer.CallOnConnStop(c)

	//关闭socket链接
	c.Conn.Close()
	//告知Writer关闭
	c.ExitChan <- true
	//将当前连接从ConnMgr中摘除
	c.TcpServer.GetConnMgr().Remove(c)

	//回收资源
	close(c.ExitChan)
	close(c.msgChan)

}

// 获取当前链接的绑定socket conn
func (c *Connection) GetTcpConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前链接模块的链接ID
func (c *Connection) GetConnId() uint32 {
	return c.ConnID
}

// 获取远程客户端的 tcp状态 Ip port
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 提供一个SendMsg方法 将我们要发送给客户端的数据，先进行封包，再发送
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	//将data进行封包 msgdaatalen | MsgID | data
	dp := NewDataPack()

	binaryMsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id =", msgId)
		return errors.New("pack error msg")
	}
	//将数据发送到客户端
	//if _, err := c.Conn.Write(binaryMsg); err != nil {
	//	fmt.Println("Write msg id ", msgId, "error:", err)
	//	return errors.New("conn Write error")
	//}
	c.msgChan <- binaryMsg
	return nil
}

// 设置链接属性
func (c *Connection) Setproperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	//添加一个链接属性
	c.property[key] = value
}

// 获取链接属性
func (c *Connection) Getproperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

// 移除链接属性
func (c *Connection) Removeproperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	//删除属性
	delete(c.property, key)
}
