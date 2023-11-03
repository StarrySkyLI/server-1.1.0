package core

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"os"
	"server-1.1.0/csvs"
	"server-1.1.0/network/utils"
	"server-1.1.0/network/ziface"
	"server-1.1.0/pb/pb"

	"sync"
	"time"
)

const (
	TASK_STATE_INIT   = 0
	TASK_STATE_DOING  = 1
	TASK_STATE_FINISH = 2
)
const (
	MOD_PLAYER     = "player"
	MOD_ICON       = "icon"
	MOD_CARD       = "card"
	MOD_UNIQUETASK = "uniquetask"
	MOD_ROLE       = "role"
	MOD_BAG        = "bag"
	MOD_WEAPON     = "weapon "
	MOD_RELICS     = "relics"
	MOD_COOK       = "cook"
	MOD_HOME       = "home"
	MOD_POOL       = "pool"
	MOD_MAP        = "map"
)

type ModBase interface {
	LoadData(player *Player)
	SaveData()
	InitData()
}

var player *Player

// 玩家对象
type Player struct {
	//玩家ID
	UserId    int32
	ModManage map[string]ModBase
	localPath string
	//当前玩家用于和客户端的连接
	Conn ziface.IConnection
	X    float32 //平面x坐标
	Y    float32 //高度
	Z    float32 //平面y坐标
	V    float32 //旋转0-360角度
}

// player id 生成器 后面生成数据库
// 应该有登录模块，由数据库查询之后再发入ID等信息
var PidGen int32 = 1
var IDLock sync.Mutex

func NewPlayer(conn ziface.IConnection) *Player {
	//生成玩家id
	IDLock.Lock()
	id := PidGen
	PidGen++
	IDLock.Unlock()
	player = new(Player)
	player.ModManage = map[string]ModBase{
		MOD_PLAYER:     new(ModPlayer),
		MOD_ICON:       new(ModIcon),
		MOD_CARD:       new(ModCard),
		MOD_UNIQUETASK: new(ModUniqueTask),
		MOD_ROLE:       new(ModRole),
		MOD_BAG:        new(ModBag),
		MOD_WEAPON:     new(ModWeapon),
		MOD_RELICS:     new(ModRelics),
		MOD_COOK:       new(ModCook),
		MOD_HOME:       new(ModHome),
		MOD_POOL:       new(ModPool),
		MOD_MAP:        new(ModMap),
	}

	//p := &Player{
	//	UserId: id,
	//	Conn:   conn,
	//
	//	X: float32(160 + rand.Intn(10)), //随机在160坐标点 基于x若干偏移
	//	Y: 0,
	//	Z: float32(140 + rand.Intn(20)),
	//	V: 0,
	//}
	player.UserId = id
	player.Conn = conn
	player.X = float32(160 + rand.Intn(10))
	player.Y = 0
	player.Z = float32(140 + rand.Intn(20))
	player.V = 0
	player.InitData()
	player.InitMod()

	return player

}

func (self *Player) InitData() {
	//path := GetServer().Config.LocalSavePath
	path := utils.GlobalObject.LocalSavePath
	_, err := os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			return
		}
	}
	self.localPath = path + fmt.Sprintf("/%d", self.UserId)
	_, err = os.Stat(self.localPath)
	if err != nil {
		err = os.Mkdir(self.localPath, os.ModePerm)
		if err != nil {
			return
		}
	}
}

func (self *Player) InitMod() {
	for _, v := range self.ModManage {
		v.LoadData(self)
	}
}

// 提供一个发送给客户端消息的方法
// 主要是将pb的protobuf数据序列化之后，再调用zinx的SendMsg方法
func (p *Player) SendMsg(msgId uint32, data proto.Message) {
	if p.Conn == nil {
		fmt.Println("connection in player is nil")
		return
	}
	//将proto Msg结构体序列化 转换为2进制
	msg, err := proto.Marshal(data)
	if err != nil {
		fmt.Println()
	}
	if err != nil {
		fmt.Println("Marshal err: ", err)

	}
	//将二进制文件通过zinx框架SendMsg将数据发送给客户端
	if p.Conn == nil {
		fmt.Printf("connection in player %d is nil", p.UserId)
		return
	}
	if err := p.Conn.SendMsg(msgId, msg); err != nil {
		fmt.Println("player send msg err")
		return
	}
	return

}

// 告知客户端玩家pid，同步已经生成的玩家id给客户端
func (p *Player) SyncPid() {
	//组建MsgID:0 的proto数据
	proto_msg := &pb.SyncPid{
		Pid: p.UserId,
	}
	p.SendMsg(1, proto_msg)
}

// 广播玩家自己的出生地点
func (p *Player) BroadCastStartPosition() {
	proto_msg := &pb.BroadCast{
		Pid: p.UserId,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}
	p.SendMsg(200, proto_msg)
}

// 玩家广播世界聊天消息
func (p *Player) WorldTalk(content string) {
	//1. 组建MsgId200 proto数据
	msg := &pb.BroadCast{
		Pid: p.UserId,
		Tp:  1, //TP 1 代表聊天广播
		Data: &pb.BroadCast_Content{
			Content: content,
		},
	}

	//2. 得到当前世界所有的在线玩家
	players := WorldMgrObj.GetAllPlayers()

	//3. 向所有的玩家发送MsgId:200消息
	for _, player := range players {
		player.SendMsg(200, msg)
	}
}

// 同步周边玩家，告知当前玩家上线，广播当前玩家位置
func (p *Player) SynvSurrounding() {
	//1 根据自己的位置，获取周围九宫格内的玩家pid
	pids := WorldMgrObj.AoiMgr.GetPidsbyPos(p.X, p.Z)
	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerByPid(int32(pid)))
	}

	//2 根据pid得到所有玩家对象
	//2.1 组建MsgID：200 proto数据
	proto_msg := &pb.BroadCast{
		Pid: p.UserId,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}
	//2.2 全部周围玩家都向格子的客户端发送200消息，让自己出现在对方视野中
	for _, player := range players {
		player.SendMsg(200, proto_msg)
	}

	//3 将周围的全部玩家的位置消息发送给当前的玩家MsgID：202 客户端（让自己看到其他玩家）
	//3.1 组建MsgID：202 proto数据
	//3.1.1制作一个pb.player slice
	players_proto_msg := make([]*pb.Player, 0, len(players))
	for _, player := range players {
		//制作一个messager player
		p := &pb.Player{
			Pid: player.UserId,
			P: &pb.Position{
				X: player.X,
				Y: player.Y,
				Z: player.Z,
				V: player.V,
			},
		}
		players_proto_msg = append(players_proto_msg, p)
	}

	//3.1.2 封装SyncPlayer protobuf 数据
	SyncPlayer_proto_msg := &pb.SyncPlayers{
		Ps: players_proto_msg[:],
	}

	//3.2 将组装好的数据发送给当前玩家客户端
	p.SendMsg(202, SyncPlayer_proto_msg)
}

// 广播并更新当前玩家坐标
func (p *Player) UpdatePos(x float32, y float32, z float32, v float32) {
	//触发消失视野和添加视野业务
	//计算旧格子gID
	oldGID := WorldMgrObj.AoiMgr.GetGidbyPos(p.X, p.Z)
	//计算新格子gID
	newGID := WorldMgrObj.AoiMgr.GetGidbyPos(x, z)

	//更新玩家的位置信息
	p.X = x
	p.Y = y
	p.Z = z
	p.V = v
	if oldGID != newGID {
		//触发gird切换
		//把pID从就的aoi格子中删除
		WorldMgrObj.AoiMgr.RemovePidfromGrid(int(p.UserId), oldGID)
		//把pID添加到新的aoi格子中去
		WorldMgrObj.AoiMgr.AddPidToGrid(int(p.UserId), newGID)

		_ = p.OnExchangeAoiGrID(oldGID, newGID)
	}

	//组装protobuf协议，发送位置给周围玩家
	proto_msg := &pb.BroadCast{
		Pid: p.UserId,
		Tp:  4, //4 移动之后的坐标信息
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//获取当前玩家周边全部玩家AOI九宫格之内的玩家
	players := p.GetSurrundingPlayers()

	//依次给每个玩家对应的客户端发送当前玩家位置更新的消息
	//向周边的每个玩家发送MsgID:200消息，移动位置更新消息
	for _, player := range players {
		player.SendMsg(200, proto_msg)
	}

}
func (p *Player) GetSurrundingPlayers() []*Player {
	//得到当前AOI九宫格内的所有玩家PID
	pids := WorldMgrObj.AoiMgr.GetPidsbyPos(p.X, p.Z)
	//将所有pid对应的Player放到Player切片中
	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerByPid(int32(pid)))
	}

	return players

}

// 玩家下线
func (p *Player) Offline() {
	//1 获取周围AOI九宫格内的玩家
	players := p.GetSurrundingPlayers()

	//2 封装MsgID:201消息
	proto_msg := &pb.SyncPid{
		Pid: p.UserId,
	}

	//3 向周围玩家发送消息
	for _, player := range players {
		player.SendMsg(201, proto_msg)
	}

	//4 世界管理器将当前玩家从AOI中摘除
	WorldMgrObj.AoiMgr.RemoveFromGridByPos(int(p.UserId), p.X, p.Z) //从格子中删除
	WorldMgrObj.RemovePlayerByid(p.UserId)
}

func (p *Player) OnExchangeAoiGrID(oldGID int, newGID int) error {
	//获取就的九宫格成员
	oldGrIDs := WorldMgrObj.AoiMgr.GetSurroundGridsByGid(oldGID)

	//为旧的九宫格成员建立哈希表,用来快速查找
	oldGrIDsMap := make(map[int]bool, len(oldGrIDs))
	for _, grID := range oldGrIDs {
		oldGrIDsMap[grID.GID] = true
	}

	//获取新的九宫格成员
	newGrIDs := WorldMgrObj.AoiMgr.GetSurroundGridsByGid(newGID)
	//为新的九宫格成员建立哈希表,用来快速查找
	newGrIDsMap := make(map[int]bool, len(newGrIDs))
	for _, grID := range newGrIDs {
		newGrIDsMap[grID.GID] = true
	}

	//------ > 处理视野消失 <-------
	offlineMsg := &pb.SyncPid{
		Pid: p.UserId,
	}

	//找到在旧的九宫格中出现,但是在新的九宫格中没有出现的格子
	leavingGrIDs := make([]*Grid, 0)
	for _, grID := range oldGrIDs {
		if _, ok := newGrIDsMap[grID.GID]; !ok {
			leavingGrIDs = append(leavingGrIDs, grID)
		}
	}

	//获取需要消失的格子中的全部玩家
	for _, grID := range leavingGrIDs {
		players := WorldMgrObj.GetPlayersByGID(grID.GID)
		for _, player := range players {
			//让自己在其他玩家的客户端中消失
			player.SendMsg(201, offlineMsg)

			//将其他玩家信息 在自己的客户端中消失
			anotherOfflineMsg := &pb.SyncPid{
				Pid: player.UserId,
			}
			p.SendMsg(201, anotherOfflineMsg)
			time.Sleep(200 * time.Millisecond)
		}
	}

	//------ > 处理视野出现 <-------

	//找到在新的九宫格内出现,但是没有在就的九宫格内出现的格子
	enteringGrIDs := make([]*Grid, 0)
	for _, grID := range newGrIDs {
		if _, ok := oldGrIDsMap[grID.GID]; !ok {
			enteringGrIDs = append(enteringGrIDs, grID)
		}
	}

	onlineMsg := &pb.BroadCast{
		Pid: p.UserId,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//获取需要显示格子的全部玩家
	for _, grID := range enteringGrIDs {
		players := WorldMgrObj.GetPlayersByGID(grID.GID)

		for _, player := range players {
			//让自己出现在其他人视野中
			player.SendMsg(200, onlineMsg)

			//让其他人出现在自己的视野中
			anotherOnlineMsg := &pb.BroadCast{
				Pid: player.UserId,
				Tp:  2,
				Data: &pb.BroadCast_P{
					P: &pb.Position{
						X: player.X,
						Y: player.Y,
						Z: player.Z,
						V: player.V,
					},
				},
			}

			time.Sleep(200 * time.Millisecond)
			p.SendMsg(200, anotherOnlineMsg)
		}
	}

	return nil

}

func (self *Player) GetMod(modName string) ModBase {
	return self.ModManage[modName]
}

func (self *Player) GetModPlayer() *ModPlayer {
	return self.ModManage[MOD_PLAYER].(*ModPlayer)
}

func (self *Player) GetModIcon() *ModIcon {
	return self.ModManage[MOD_ICON].(*ModIcon)
}

func (self *Player) GetModCard() *ModCard {
	return self.ModManage[MOD_CARD].(*ModCard)
}

func (self *Player) GetModUniqueTask() *ModUniqueTask {
	return self.ModManage[MOD_UNIQUETASK].(*ModUniqueTask)
}

func (self *Player) GetModRole() *ModRole {
	return self.ModManage[MOD_ROLE].(*ModRole)
}

func (self *Player) GetModBag() *ModBag {
	return self.ModManage[MOD_BAG].(*ModBag)
}

func (self *Player) GetModWeapon() *ModWeapon {
	return self.ModManage[MOD_WEAPON].(*ModWeapon)
}

func (self *Player) GetModRelics() *ModRelics {
	return self.ModManage[MOD_RELICS].(*ModRelics)
}

func (self *Player) GetModCook() *ModCook {
	return self.ModManage[MOD_COOK].(*ModCook)
}

func (self *Player) GetModHome() *ModHome {
	return self.ModManage[MOD_HOME].(*ModHome)
}

func (self *Player) GetModPool() *ModPool {
	return self.ModManage[MOD_POOL].(*ModPool)
}

func (self *Player) GetModMap() *ModMap {
	return self.ModManage[MOD_MAP].(*ModMap)
}

// 基础信息
func (self *Player) HandleBase() {
	for {
		fmt.Println("当前处于基础信息界面,请选择操作：0返回1查询信息2设置名字3设置签名4头像5名片6设置生日")
		var action int
		fmt.Scan(&action)
		switch action {
		case 0:
			return
		case 1:
			self.HandleBaseGetInfo()
		case 2:
			self.HandleBagSetName()
		case 3:
			self.HandleBagSetSign()
		case 4:
			self.HandleBagSetIcon()
		case 5:
			self.HandleBagSetCard()
		case 6:
			self.HandleBagSetBirth()
		}
	}
}
func (self *Player) HandleBagSetSign() {
	fmt.Println("请输入签名:")
	var sign string
	fmt.Scan(&sign)
	self.RecvSetSign(sign)
}
func (self *Player) HandleBagSetName() {
	fmt.Println("请输入名字:")
	var name string
	fmt.Scan(&name)
	self.RecvSetName(name)
}
func (self *Player) HandleBaseGetInfo() {
	fmt.Println("名字:", self.GetMod(MOD_PLAYER).(*ModPlayer).Name)
	fmt.Println("等级:", self.GetMod(MOD_PLAYER).(*ModPlayer).PlayerLevel)
	fmt.Println("大世界等级:", self.GetMod(MOD_PLAYER).(*ModPlayer).WorldLevelNow)
	if self.GetMod(MOD_PLAYER).(*ModPlayer).Sign == "" {
		fmt.Println("签名:", "未设置")
	} else {
		fmt.Println("签名:", self.GetMod(MOD_PLAYER).(*ModPlayer).Sign)
	}

	if self.GetMod(MOD_PLAYER).(*ModPlayer).Icon == 0 {
		fmt.Println("头像:", "未设置")
	} else {
		fmt.Println("头像:", csvs.GetItemConfig(self.GetMod(MOD_PLAYER).(*ModPlayer).Icon), self.GetMod(MOD_PLAYER).(*ModPlayer).Icon)
	}

	if self.GetMod(MOD_PLAYER).(*ModPlayer).Card == 0 {
		fmt.Println("名片:", "未设置")
	} else {
		fmt.Println("名片:", csvs.GetItemConfig(self.GetMod(MOD_PLAYER).(*ModPlayer).Card), self.GetMod(MOD_PLAYER).(*ModPlayer).Card)
	}

	if self.GetMod(MOD_PLAYER).(*ModPlayer).Birth == 0 {
		fmt.Println("生日:", "未设置")
	} else {
		fmt.Println("生日:", self.GetMod(MOD_PLAYER).(*ModPlayer).Birth/100, "月", self.GetMod(MOD_PLAYER).(*ModPlayer).Birth%100, "日")
	}
}

func (self *Player) HandleBagSetIcon() {
	for {
		fmt.Println("当前处于基础信息--头像界面,请选择操作：0返回1查询头像背包2设置头像")
		var action int
		fmt.Scan(&action)
		switch action {
		case 0:
			return
		case 1:
			self.HandleBagSetIconGetInfo()
		case 2:
			self.HandleBagSetIconSet()
		}
	}
}

func (self *Player) HandleBagSetIconGetInfo() {
	fmt.Println("当前拥有头像如下:")
	for _, v := range self.GetModIcon().IconInfo {
		config := csvs.GetItemConfig(v.IconId)
		if config != nil {
			fmt.Println(config.ItemName, ":", config.ItemId)
		}
	}
}

func (self *Player) HandleBagSetIconSet() {
	fmt.Println("请输入头像id:")
	var icon int
	fmt.Scan(&icon)
	self.RecvSetIcon(icon)
}

func (self *Player) HandleBagSetCard() {
	for {
		fmt.Println("当前处于基础信息--名片界面,请选择操作：0返回1查询名片背包2设置名片")
		var action int
		fmt.Scan(&action)
		switch action {
		case 0:
			return
		case 1:
			self.HandleBagSetCardGetInfo()
		case 2:
			self.HandleBagSetCardSet()
		}
	}
}

func (self *Player) HandleBagSetCardGetInfo() {
	fmt.Println("当前拥有名片如下:")
	for _, v := range self.GetModCard().CardInfo {
		config := csvs.GetItemConfig(v.CardId)
		if config != nil {
			fmt.Println(config.ItemName, ":", config.ItemId)
		}
	}
}

func (self *Player) HandleBagSetCardSet() {
	fmt.Println("请输入名片id:")
	var card int
	fmt.Scan(&card)
	self.RecvSetCard(card)
}

func (self *Player) HandleBagSetBirth() {
	if self.GetMod(MOD_PLAYER).(*ModPlayer).Birth > 0 {
		fmt.Println("已设置过生日!")
		return
	}
	fmt.Println("生日只能设置一次，请慎重填写,输入月:")
	var month, day int
	fmt.Scan(&month)
	fmt.Println("请输入日:")
	fmt.Scan(&day)
	self.GetMod(MOD_PLAYER).(*ModPlayer).SetBirth(month*100 + day)
}

// 背包
func (self *Player) HandleBag() {
	for {
		fmt.Println("当前处于基础信息界面,请选择操作：0返回1增加物品2扣除物品3使用物品4升级七天神像(风)")
		var action int
		fmt.Scan(&action)
		switch action {
		case 0:
			return
		case 1:
			self.HandleBagAddItem()
		case 2:
			self.HandleBagRemoveItem()
		case 3:
			self.HandleBagUseItem()
		case 4:
			self.HandleBagWindStatue()
		}
	}
}

// 抽卡
func (self *Player) HandlePool() {
	for {
		fmt.Println("当前处于模拟抽卡界面,请选择操作：0返回1角色信息2十连抽(入包)3单抽(可选次数,入包)" +
			"4五星爆点测试5十连多黄测试6视频原版函数(30秒)7单抽(仓检版,独宠一人)8单抽(仓检版,雨露均沾)")
		var action int
		fmt.Scan(&action)
		switch action {
		case 0:
			return
		case 1:
			self.GetModRole().HandleSendRoleInfo(self)
		case 2:
			self.GetModPool().HandleUpPoolTen(self)
		case 3:
			fmt.Println("请输入抽卡次数,最大值1亿(最大耗时约30秒):")
			var times int
			fmt.Scan(&times)
			self.GetModPool().HandleUpPoolSingle(times, self)
		case 4:
			fmt.Println("请输入抽卡次数,最大值1亿(最大耗时约30秒):")
			var times int
			fmt.Scan(&times)
			self.GetModPool().HandleUpPoolTimesTest(times)
		case 5:
			fmt.Println("请输入抽卡次数,最大值1亿(最大耗时约30秒):")
			var times int
			fmt.Scan(&times)
			self.GetModPool().HandleUpPoolFiveTest(times)
		case 6:
			self.GetModPool().DoUpPool()
		case 7:
			fmt.Println("请输入抽卡次数,最大值1亿(最大耗时约30秒):")
			var times int
			fmt.Scan(&times)
			self.GetModPool().HandleUpPoolSingleCheck1(times, self)
		case 8:
			fmt.Println("请输入抽卡次数,最大值1亿(最大耗时约30秒):")
			var times int
			fmt.Scan(&times)
			self.GetModPool().HandleUpPoolSingleCheck2(times, self)
		}
	}
}

func (self *Player) HandleBagAddItem() {
	itemId := 0
	itemNum := 0
	fmt.Println("物品ID")
	fmt.Scan(&itemId)
	fmt.Println("物品数量")
	fmt.Scan(&itemNum)
	self.GetModBag().AddItem(itemId, int64(itemNum))
}

func (self *Player) HandleBagRemoveItem() {
	itemId := 0
	itemNum := 0
	fmt.Println("物品ID")
	fmt.Scan(&itemId)
	fmt.Println("物品数量")
	fmt.Scan(&itemNum)
	self.GetModBag().RemoveItemToBag(itemId, int64(itemNum))
}

func (self *Player) HandleBagUseItem() {
	itemId := 0
	itemNum := 0
	fmt.Println("物品ID")
	fmt.Scan(&itemId)
	fmt.Println("物品数量")
	fmt.Scan(&itemNum)
	self.GetModBag().UseItem(itemId, int64(itemNum))
}

func (self *Player) HandleBagWindStatue() {
	fmt.Println("开始升级七天神像")
	self.GetModMap().UpStatue(1)
	self.GetModRole().CalHpPool()
}

// 地图
func (self *Player) HandleMap() {
	fmt.Println("向着星辰与深渊,欢迎来到冒险家协会！")
	for {
		fmt.Println("请选择互动地图1蒙德2璃月1001深入风龙废墟2001无妄引咎密宫")
		var action int
		fmt.Scan(&action)
		switch action {
		case 0:
			return
		default:
			self.HandleMapIn(action)
		}
	}
}

func (self *Player) HandleMapIn(mapId int) {

	config := csvs.ConfigMapMap[mapId]
	if config == nil {
		fmt.Println("无法识别的地图")
		return
	}
	self.GetModMap().RefreshByPlayer(mapId)
	for {
		self.GetModMap().GetEventList(config)
		fmt.Println("请选择触发事件Id(0返回)")
		var action int
		fmt.Scan(&action)
		switch action {
		case 0:
			return
		default:
			eventConfig := csvs.ConfigMapEventMap[action]
			if eventConfig == nil {
				fmt.Println("无法识别的事件")
				break
			}
			self.GetModMap().SetEventState(mapId, eventConfig.EventId, csvs.EVENT_END, self)
		}
	}
}

func (self *Player) HandleRelics() {
	for {
		fmt.Println("当前处于圣遗物界面，选择功能0返回1强化测试2满级圣遗物3极品头测试")
		var action int
		fmt.Scan(&action)
		switch action {
		case 0:
			return
		case 1:
			self.GetModRelics().RelicsUp(self)
		case 2:
			self.GetModRelics().RelicsTop(self)
		case 3:
			self.GetModRelics().RelicsTestBest(self)
		default:
			fmt.Println("无法识别在操作")
		}
	}
}

func (self *Player) HandleRole() {
	for {
		fmt.Println("当前处于角色界面，选择功能0返回1查询2穿戴圣遗物3卸下圣遗物4穿戴武器5卸下武器")
		var action int
		fmt.Scan(&action)
		switch action {
		case 0:
			return
		case 1:
			self.GetModRole().HandleSendRoleInfo(self)
		case 2:
			self.HandleWearRelics()
		case 3:
			self.HandleTakeOffRelics()
		case 4:
			self.HandleWearWeapon()
		case 5:
			self.HandleTakeOffWeapon()
		default:
			fmt.Println("无法识别在操作")
		}
	}
}

func (self *Player) HandleWearRelics() {
	for {
		fmt.Println("输入操作的目标英雄Id:,0返回")
		var roleId int
		fmt.Scan(&roleId)

		if roleId == 0 {
			return
		}

		RoleInfo := self.GetModRole().RoleInfo[roleId]
		if RoleInfo == nil {
			fmt.Println("英雄不存在")
			continue
		}

		RoleInfo.ShowInfo(self)
		fmt.Println("输入需要穿戴的圣遗物key:,0返回")
		var relicsKey int
		fmt.Scan(&relicsKey)
		if relicsKey == 0 {
			return
		}
		relics := self.GetModRelics().RelicsInfo[relicsKey]
		if relics == nil {
			fmt.Println("圣遗物不存在")
			continue
		}
		self.GetModRole().WearRelics(RoleInfo, relics, self)
	}
}

func (self *Player) HandleTakeOffRelics() {
	for {
		fmt.Println("输入操作的目标英雄Id:,0返回")
		var roleId int
		fmt.Scan(&roleId)

		if roleId == 0 {
			return
		}

		RoleInfo := self.GetModRole().RoleInfo[roleId]
		if RoleInfo == nil {
			fmt.Println("英雄不存在")
			continue
		}

		RoleInfo.ShowInfo(self)
		fmt.Println("输入需要卸下的圣遗物key:,0返回")
		var relicsKey int
		fmt.Scan(&relicsKey)
		if relicsKey == 0 {
			return
		}
		relics := self.GetModRelics().RelicsInfo[relicsKey]
		if relics == nil {
			fmt.Println("圣遗物不存在")
			continue
		}
		self.GetModRole().TakeOffRelics(RoleInfo, relics, self)
	}
}

func (self *Player) HandleWeapon() {
	for {
		fmt.Println("当前处于武器界面，选择功能0返回1强化测试2突破测试3精炼测试")
		var action int
		fmt.Scan(&action)
		switch action {
		case 0:
			return
		case 1:
			self.HandleWeaponUp()
		case 2:
			self.HandleWeaponStarUp()
		case 3:
			self.HandleWeaponRefineUp()
		default:
			fmt.Println("无法识别在操作")
		}
	}
}

func (self *Player) HandleWeaponUp() {
	for {
		fmt.Println("输入操作的目标武器keyId:,0返回")
		for _, v := range self.GetModWeapon().WeaponInfo {
			fmt.Println(fmt.Sprintf("武器keyId:%d,等级:%d,突破等级:%d,精炼:%d",
				v.KeyId, v.Level, v.StarLevel, v.RefineLevel))
		}
		var weaponKeyId int
		fmt.Scan(&weaponKeyId)
		if weaponKeyId == 0 {
			return
		}
		self.GetModWeapon().WeaponUp(weaponKeyId, self)
	}
}

func (self *Player) HandleWeaponStarUp() {
	for {
		fmt.Println("输入操作的目标武器keyId:,0返回")
		for _, v := range self.GetModWeapon().WeaponInfo {
			fmt.Println(fmt.Sprintf("武器keyId:%d,等级:%d,突破等级:%d,精炼:%d",
				v.KeyId, v.Level, v.StarLevel, v.RefineLevel))
		}
		var weaponKeyId int
		fmt.Scan(&weaponKeyId)
		if weaponKeyId == 0 {
			return
		}
		self.GetModWeapon().WeaponUpStar(weaponKeyId, self)
	}
}

func (self *Player) HandleWeaponRefineUp() {
	for {
		fmt.Println("输入操作的目标武器keyId:,0返回")
		for _, v := range self.GetModWeapon().WeaponInfo {
			fmt.Println(fmt.Sprintf("武器keyId:%d,等级:%d,突破等级:%d,精炼:%d",
				v.KeyId, v.Level, v.StarLevel, v.RefineLevel))
		}
		var weaponKeyId int
		fmt.Scan(&weaponKeyId)
		if weaponKeyId == 0 {
			return
		}
		for {
			fmt.Println("输入作为材料的武器keyId:,0返回")
			var weaponTargetKeyId int
			fmt.Scan(&weaponTargetKeyId)
			if weaponTargetKeyId == 0 {
				return
			}
			self.GetModWeapon().WeaponUpRefine(weaponKeyId, weaponTargetKeyId, self)
		}
	}
}

func (self *Player) HandleWearWeapon() {
	for {
		fmt.Println("输入操作的目标英雄Id:,0返回")
		var roleId int
		fmt.Scan(&roleId)

		if roleId == 0 {
			return
		}

		RoleInfo := self.GetModRole().RoleInfo[roleId]
		if RoleInfo == nil {
			fmt.Println("英雄不存在")
			continue
		}

		RoleInfo.ShowInfo(self)
		fmt.Println("输入需要穿戴的武器key:,0返回")
		var weaponKey int
		fmt.Scan(&weaponKey)
		if weaponKey == 0 {
			return
		}
		weaponInfo := self.GetModWeapon().WeaponInfo[weaponKey]
		if weaponInfo == nil {
			fmt.Println("武器不存在")
			continue
		}
		self.GetModRole().WearWeapon(RoleInfo, weaponInfo, self)
		RoleInfo.ShowInfo(self)
	}
}

func (self *Player) HandleTakeOffWeapon() {
	for {
		fmt.Println("输入操作的目标英雄Id:,0返回")
		var roleId int
		fmt.Scan(&roleId)

		if roleId == 0 {
			return
		}

		RoleInfo := self.GetModRole().RoleInfo[roleId]
		if RoleInfo == nil {
			fmt.Println("英雄不存在")
			continue
		}

		RoleInfo.ShowInfo(self)
		fmt.Println("输入需要卸下的武器key:,0返回")
		var weaponKey int
		fmt.Scan(&weaponKey)
		if weaponKey == 0 {
			return
		}
		weapon := self.GetModWeapon().WeaponInfo[weaponKey]
		if weapon == nil {
			fmt.Println("武器不存在")
			continue
		}
		self.GetModRole().TakeOffWeapon(RoleInfo, weapon, self)
		RoleInfo.ShowInfo(self)
	}
} //对外接口
func (self *Player) RecvSetIcon(iconId int) {
	self.GetMod(MOD_PLAYER).(*ModPlayer).SetIcon(iconId)
}

func (self *Player) RecvSetCard(cardId int) {
	self.GetMod(MOD_PLAYER).(*ModPlayer).SetCard(cardId)
}

func (self *Player) RecvSetName(name string) {
	self.GetMod(MOD_PLAYER).(*ModPlayer).SetName(name)
}

func (self *Player) RecvSetSign(sign string) {
	self.GetMod(MOD_PLAYER).(*ModPlayer).SetSign(sign)
}

func (self *Player) ReduceWorldLevel() {
	self.GetMod(MOD_PLAYER).(*ModPlayer).ReduceWorldLevel()
}

func (self *Player) ReturnWorldLevel() {
	self.GetMod(MOD_PLAYER).(*ModPlayer).ReturnWorldLevel()
}

func (self *Player) SetBirth(birth int) {
	self.GetMod(MOD_PLAYER).(*ModPlayer).SetBirth(birth)
}

func (self *Player) SetShowCard(showCard []int) {
	self.GetMod(MOD_PLAYER).(*ModPlayer).SetShowCard(showCard, self)
}

func (self *Player) SetShowTeam(showRole []int) {
	self.GetMod(MOD_PLAYER).(*ModPlayer).SetShowTeam(showRole, self)
}

func (self *Player) SetHideShowTeam(isHide int) {
	self.GetMod(MOD_PLAYER).(*ModPlayer).SetHideShowTeam(isHide, self)
}

func (self *Player) SetEventState(state int) {
	//self.ModMap.SetEventState(state, self)
}
