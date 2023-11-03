package apis

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"strconv"

	"server-1.1.0/core"
	"server-1.1.0/network/ziface"
	"server-1.1.0/network/znet"
	"server-1.1.0/pb/pb"
)

// 玩家开游戏路由
type GamesApi struct {
	znet.BaseRouter
}

func (g *GamesApi) Handle(request ziface.IRequest) {
	//解析客户端发来的proto协议
	//1 解析客户端传递进来的proto协议
	proto_msg := &pb.Game{}
	err := proto.Unmarshal(request.GetData(), proto_msg)
	if err != nil {
		fmt.Println("Talk Unmarshal error ", err)
		return
	}

	//2 得到当前连接pid
	pid, err := request.GetConnection().Getproperty("pid")
	if err != nil {
		fmt.Println(err)
	}
	//3 根据pid得到当前玩家对应的player对象
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))
	var modChoose int
	modChoose, _ = strconv.Atoi(proto_msg.Content)
	switch modChoose {
	case 1:
		player.HandleBase()
	case 2:
		player.HandleBag()
	case 3:
		player.HandlePool()
	case 4:
		player.HandleMap()
	case 5:
		player.HandleRelics()
	case 6:
		player.HandleRole()
	case 7:
		player.HandleWeapon()
	case 8:
		for _, v := range player.ModManage {
			v.SaveData()
		}
	}

}
