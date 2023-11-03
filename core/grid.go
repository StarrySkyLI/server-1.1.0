package core

import (
	"fmt"
	"sync"
)

// 一个AOI的地图中的格子类型
type Grid struct {
	//格子ID
	GID int
	//格子左边界坐标
	MinX int
	//格子右边界坐标
	MaxX int
	//格子上边界坐标
	MinY int
	//格子下边界坐标
	MaxY int
	//当前格子内的玩家或者物体成员ID
	playerIDs map[int]bool
	//playerIDs的保护map的锁
	pIDLock sync.RWMutex
}

// 初始化一个格子
func NewGrid(gID, minX, maxX, minY, maxY int) *Grid {
	return &Grid{
		GID:       gID,
		MinX:      minX,
		MaxY:      maxY,
		MinY:      minY,
		MaxX:      maxX,
		playerIDs: make(map[int]bool),
	}

}

// 向当前格子中添加一个玩家
func (g *Grid) Add(playerID int) {
	g.pIDLock.Lock()
	defer g.pIDLock.Unlock()

	g.playerIDs[playerID] = true
}

// 从格子中删除一个玩家
func (g *Grid) Remove(playerID int) {
	g.pIDLock.Lock()
	defer g.pIDLock.Unlock()

	delete(g.playerIDs, playerID)
}

// 得到当前格子中所有的玩家
func (g *Grid) GetPlayerIDs() (playerIDs []int) {
	g.pIDLock.RLock()
	defer g.pIDLock.RUnlock()
	for k, _ := range g.playerIDs {
		playerIDs = append(playerIDs, k)
	}
	return

}

// 调试使用---打印信息方法打印出格子基本信息
func (g *Grid) String() string {
	return fmt.Sprintf("Grid id: %d,minX:%d,maxX:%d,minY:%d,maxY:%d,playerIDs:%v",
		g.GID, g.MinX, g.MaxX, g.MinY, g.MaxY, g.playerIDs)
}
