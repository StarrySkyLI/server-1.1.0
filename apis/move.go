package apis

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"server-1.1.0/core"
	"server-1.1.0/network/ziface"
	"server-1.1.0/network/znet"
	"server-1.1.0/pb/pb"
)

// 玩家移动路由
type MoveApi struct {
	znet.BaseRouter
}

func (m *MoveApi) Handle(request ziface.IRequest) {
	//解析客户端发来的proto协议
	proto_msg := &pb.Position{}
	err := proto.Unmarshal(request.GetData(), proto_msg)
	if err != nil {
		fmt.Println("move : position Unmarshal error", err)
		return
	}
	//得到当前发送位置的是哪个玩家
	pid, err := request.GetConnection().Getproperty("pid")
	if err != nil {
		fmt.Println("GetProperty pid error ", err)
		request.GetConnection().Stop()
		return
	}
	fmt.Printf("Player pid =%d ,move(%f,%f,%f，%f)\n", pid, proto_msg.X, proto_msg.Y, proto_msg.Z, proto_msg.V)

	//给其他玩家进行当前玩家的位置消息广播
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))
	//广播并更新当前玩家坐标
	player.UpdatePos(proto_msg.X, proto_msg.Y, proto_msg.Z, proto_msg.V)
}
