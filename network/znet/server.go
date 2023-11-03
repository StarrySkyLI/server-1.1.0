package znet

import (
	"fmt"
	"regexp"
	"server-1.1.0/network/utils"
	"server-1.1.0/network/ziface"
	"sync"

	"net"
)

// 实现层
// iServer的接口实现，定义一个server的服务器模块
type Server struct {

	//服务器名字
	Name string
	//服务器绑定ip版本
	IPVersion string
	//监听的ip
	IP string
	//监听的端口
	Port int
	//当前的server的消息管理模块，原来绑定MsgId和对应的处理业务API关系
	MsgHandler ziface.IMsgHanle
	//该server的连接管理器
	ConnMgr     ziface.IConnManager
	Lock        *sync.RWMutex
	BanWordBase []string //配置生成
	Wait        sync.WaitGroup

	// =======================
	//新增两个hook函数原型

	//该Server的连接创建时Hook函数 OnConnStart
	OnConnStart func(conn ziface.IConnection)
	//该Server的连接断开时的Hook函数 OnConnStop
	OnConnStop func(conn ziface.IConnection)

	// =======================
}

func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {

	s.MsgHandler.AddRouter(msgID, router)
	fmt.Println("add Router Succ")
}
func (self *Server) IsBanWord(txt string) bool {
	self.Lock.RLock()
	defer self.Lock.RUnlock()
	for _, v := range self.BanWordBase {
		match, _ := regexp.MatchString(v, txt)
		if match {
			fmt.Println("发现违禁词:", v)
		}
		if match {
			return match
		}
	}
	return false
}
func (self *Server) UpdateBanWord(banWord []string) {
	self.Lock.Lock()
	defer self.Lock.Unlock()
	self.BanWordBase = banWord
}

// 启动网络服务
func (s *Server) Start() {
	fmt.Printf("[Start] server name : %s,listener at Ip: %s ,Port:%d is starting\n",
		utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)
	fmt.Printf("[Zinx] version %s, MaxConn:%d ,MaxPacketSize:%d\n",
		utils.GlobalObject.Version, utils.GlobalObject.MaxConn, utils.GlobalObject.MaxPacketSize)

	go func() {
		//0 开启开启消息队列及工作池
		s.MsgHandler.StartWorkerPool()
		//1 获取TCP的addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err", err)
			return
		}
		//2 监听服务器地址
		lisenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}
		fmt.Println("start Zinx server success", s.Name, "succ,Listenning...")
		var cid uint32
		cid = 0
		//3 阻塞的等待客户端链接，处理客户端链接业务（读写）
		for {
			//如果有客户端链接过来，阻塞会返回
			conn, err := lisenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}

			//设置最大连接个数的判断，如果超过最大连接，那么则关闭此新的连接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				//todo 给客户端响应一个错误包
				fmt.Println("Too Many Connection MaxConn =", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}
			//将处理新链接的业务方法和conn进行绑定，得到我们的链接模块
			dealConn := NewConnection(s, conn, cid, s.MsgHandler)
			cid++
			//启动当前的链接业务处理
			go dealConn.Start()
		}
	}()

}
func (s *Server) Stop() {
	// 将一些服务器资源状态，已经开辟的链接消息停止
	fmt.Println("[STOP] Zinx server name", s.Name)
	s.ConnMgr.ClearConn()
}
func (s *Server) Serve() {
	//启动server的服务功能
	s.Start()

	//todo 做一些服务器之后的额外业务
	//阻塞状态
	select {}
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr

}

/*
  创建一个服务器句柄
*/
// 初始化server模块的方法
func NewServer() ziface.IServer {
	s := &Server{
		Name:       utils.GlobalObject.Name,
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TcpPort,
		MsgHandler: NewMsgHandle(),
		ConnMgr:    NewManager(),
	}
	return s
}

// 注册OnConnStart钩子函数的方法
func (s *Server) SetOnConnStart(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// 注册OnConnStop钩子函数的方法
func (s *Server) SetOnConnStop(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// 调用OnConnStart钩子函数的方法
func (s *Server) CallOnConnStart(connection ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---->Call onConnStart()...")
		s.OnConnStart(connection)
	}
}

// 调用OnConnStop钩子函数的方法
func (s *Server) CallOnConnStop(connection ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---->Call onConnStop()...")
		s.OnConnStop(connection)
	}
}

var server *Server

func GetServer() *Server {
	if server == nil {
		server = new(Server)
		server.Lock = new(sync.RWMutex)
	}
	return server
}
func (self *Server) AddGo() {
	self.Wait.Add(1)
}

func (self *Server) GoDone() {
	self.Wait.Done()
}
