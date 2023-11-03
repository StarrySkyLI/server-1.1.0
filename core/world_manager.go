package core

import "sync"

// 当前游戏的世界总管理模块
type WorldManager struct {
	AoiMgr  *AOIManager       //当前世界地图的AOI规划管理器
	Players map[int32]*Player //当前在线的玩家集合

	pLock sync.RWMutex //保护Players的互斥读写机制
}

// 提供一个对外世界管理模块句柄（全局）
var WorldMgrObj *WorldManager

// 提供WorldManager 初始化方法
func init() {
	WorldMgrObj = &WorldManager{
		AoiMgr:  NewAOIManager(AOI_MIN_X, AOI_MAX_X, AOI_CNTS_X, AOI_MIN_Y, AOI_MAX_Y, AOI_CNTS_Y),
		Players: make(map[int32]*Player),
	}
}

// 提供添加一个玩家的的功能，将玩家添加进玩家信息表Players
func (wm *WorldManager) AddPlayer(player *Player) {
	wm.pLock.Lock()
	wm.Players[player.UserId] = player
	wm.pLock.Unlock()
	wm.AoiMgr.AddtoGridByPos(int(player.UserId), player.X, player.Z)
}

// 从玩家信息表中移除一个玩家
func (wm *WorldManager) RemovePlayerByid(pid int32) {
	player := wm.Players[pid]
	//AOI中删除
	wm.AoiMgr.RemoveFromGridByPos(int(pid), player.X, player.Z)

	wm.pLock.Lock()
	delete(wm.Players, pid)
	wm.pLock.Unlock()
}

// 通过玩家ID 获取对应玩家信息
func (wm *WorldManager) GetPlayerByPid(pid int32) *Player {
	wm.pLock.RLock()
	defer wm.pLock.RUnlock()
	return wm.Players[pid]
}

// 获取所有在线玩家的信息
func (wm *WorldManager) GetAllPlayers() []*Player {
	wm.pLock.RLock()
	defer wm.pLock.RUnlock()

	players := make([]*Player, 0)
	//添加进切片
	for _, p := range wm.Players {
		players = append(players, p)
	}
	return players
}

// 获取指定gID中的所有player信息
func (wm *WorldManager) GetPlayersByGID(gID int) []*Player {
	//通过gID获取 对应 格子中的所有pID
	pIDs := wm.AoiMgr.grids[gID].GetPlayerIDs()

	//通过pID找到对应的player对象
	players := make([]*Player, 0, len(pIDs))
	wm.pLock.RLock()
	for _, pID := range pIDs {
		players = append(players, wm.Players[int32(pID)])
	}
	wm.pLock.RUnlock()

	return players
}
