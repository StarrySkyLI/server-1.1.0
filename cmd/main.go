package main

import (
	"fmt"
	"server-1.1.0/apis"
	"server-1.1.0/core"
	"server-1.1.0/csvs"
	"server-1.1.0/network/ziface"
	"server-1.1.0/network/znet"
	"server-1.1.0/pb/pb"
)

// 当前客户端建立连接后的hook函数
func OnConnectionAdd(conn ziface.IConnection) {
	//创建player
	player := core.NewPlayer(conn)
	name := player.GetModPlayer().Name
	msg := &pb.Game{
		Content: name + "请选择功能：1基础信息2背包3角色(八重神子UP池)4地图5圣遗物6角色7武器8存储数据",
	}
	player.SendMsg(4, msg)

	//给客户端发送MsgID：1的消息 :同步当前player的id给客户端
	player.SyncPid()
	//给客户端发送MsgID：200的消息：同步初始化位置
	player.BroadCastStartPosition()
	//将当前新上线玩家添加到worldManager中
	core.WorldMgrObj.AddPlayer(player)
	//将该连接绑定一个pid玩家ID的属性
	conn.Setproperty("pid", player.UserId)
	//同步周边玩家，告知当前玩家上线，广播当前玩家位置
	player.SynvSurrounding()

	fmt.Println("===>player pid ", player.UserId, " is arrived<=====")
}

// 当前客户端断开连接后的hook函数
func OnConnectionLost(conn ziface.IConnection) {
	//获取当前连接的绑定的Pid
	pid, _ := conn.Getproperty("pid")

	//根据pid获取对应的玩家对象
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))

	//触发玩家下线业务
	player.Offline()

	fmt.Println("====> Player ", pid, " left =====")
}
func main() {
	//创建zinx server句柄
	s := znet.NewServer()
	csvs.CheckLoadCsv()

	//连接创建和销毁的HOOK钩子函数
	s.SetOnConnStart(OnConnectionAdd)
	s.SetOnConnStop(OnConnectionLost)

	//注册一些路由业务
	s.AddRouter(2, &apis.WorldChatApi{})
	s.AddRouter(3, &apis.MoveApi{})
	s.AddRouter(4, &apis.GamesApi{})

	//启动服务
	s.Serve()

}
