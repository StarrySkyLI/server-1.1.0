package core

import "fmt"

// 定义一些AOI的边界值
const (
	AOI_MIN_X  int = 85
	AOI_MAX_X  int = 410
	AOI_CNTS_X int = 10
	AOI_MIN_Y  int = 75
	AOI_MAX_Y  int = 400
	AOI_CNTS_Y int = 20
)

// AOI区域管理模块
type AOIManager struct {
	//区域左边界坐标
	MinX int
	//区域右边界坐标
	MaxX int
	//x方向格子的数量
	CntsX int
	//区域上边界坐标
	MinY int
	//区域下边界坐标
	MaxY int
	//y方向的格子数量
	CntsY int
	//当前区域中都有哪些格子，key=格子ID， value=格子对象
	grids map[int]*Grid
}

// 初始化一个AOI区域管理模块
func NewAOIManager(minX, maxX, cntsX, minY, maxY, cntsY int) *AOIManager {
	aoiMgr := &AOIManager{
		MinX:  minX,
		MaxX:  maxX,
		CntsX: cntsX,
		MinY:  minY,
		MaxY:  maxY,
		CntsY: cntsY,
		grids: make(map[int]*Grid),
	}
	//给AOI初始化区域的格子所有格子进行编号和初始化
	for y := 0; y < cntsY; y++ {
		for x := 0; x < cntsX; x++ {
			//计算格子ID 根据x，y编号
			//格子编号： id= idy*cntX + idx
			gid := y*cntsX + x

			//初始化gid格子
			aoiMgr.grids[gid] = NewGrid(gid,
				aoiMgr.MinX+x*aoiMgr.gridWidth(),
				aoiMgr.MinX+(x+1)*aoiMgr.gridWidth(),
				aoiMgr.MinY+y*aoiMgr.gridLength(),
				aoiMgr.MinY+(y+1)*aoiMgr.gridLength(),
			)

		}
	}

	return aoiMgr

}

// 得到每个格子在X轴方向的宽度
func (m *AOIManager) gridWidth() int {
	return (m.MaxX - m.MinX) / m.CntsX
}

// 得到每个格子在Y轴方向的高度
func (m *AOIManager) gridLength() int {
	return (m.MaxY - m.MinY) / m.CntsY
}

// 打印格子信息
func (m *AOIManager) String() string {
	//打印AOIManager信息
	s := fmt.Sprintf("AOIManagr:\nminX:%d, maxX:%d, cntsX:%d, minY:%d, maxY:%d, cntsY:%d\n Grids in AOI Manager:\n",
		m.MinX, m.MaxX, m.CntsX, m.MinY, m.MaxY, m.CntsY)
	//打印全部格子信息
	for _, grid := range m.grids {
		s += fmt.Sprintln(grid)
	}

	return s
}

// 根据格子GID得到周边九宫格格子的ID集合
func (m *AOIManager) GetSurroundGridsByGid(gID int) (grids []*Grid) {
	//判断gID是否在AOIManager中
	if _, ok := m.grids[gID]; !ok {
		return
	}

	//初始化grids返回值切片，将当前gid添加到九宫格中
	grids = append(grids, m.grids[gID])

	//根据gID得到当前格子所在的X轴编号--idx =id%nx
	idx := gID % m.CntsX
	//判断当前idx左边是否还有格子，如果有放在gidsX集合中
	if idx > 0 {
		grids = append(grids, m.grids[gID-1])
	}

	//判断当前的idx右边是否还有格子，如果有放在gidsX集合中
	if idx < m.CntsX-1 {
		grids = append(grids, m.grids[gID+1])

	}
	// 将x轴当前的格子取出遍历，再分别得到每个格子上下是否还有格子
	//得到当前x轴格子的ID集合
	gidsX := make([]int, 0, len(grids))
	for _, v := range grids {
		gidsX = append(gidsX, v.GID)
	}
	//遍历gidsX集合中每个格子的gid
	for _, v := range gidsX {
		//得到当前格子id的y轴编号 idy=id/ny
		idy := v / m.CntsY

		//判断当前的id上边是否还有格子
		if idy > 0 {
			grids = append(grids, m.grids[v-m.CntsX])
		}

		//判断当前的id下边是否还有格子
		if idy < m.CntsY-1 {
			grids = append(grids, m.grids[v+m.CntsX])
		}

	}

	return
}

// 通过横纵坐标得到当前GID格子编号
func (m *AOIManager) GetGidbyPos(x, y float32) int {
	idx := (int(x) - m.MinX) / m.gridWidth()
	idy := (int(y) - m.MinY) / m.gridLength()

	return idy*m.CntsX + idx

}

// 通过横纵坐标得到周边九宫格内全部的PlayerIDs
func (m *AOIManager) GetPidsbyPos(x, y float32) (playerIDs []int) {
	//得到当前玩家的GID格子id
	gID := m.GetGidbyPos(x, y)

	//通过GID得到周边九宫格信息
	grids := m.GetSurroundGridsByGid(gID)

	//将九宫格信息里的全部Player的id累加到playerIDs
	for _, grid := range grids {
		playerIDs = append(playerIDs, grid.GetPlayerIDs()...)
		//fmt.Println("===>grid ID :%d,pid :v% ===", grid.GID, grid.GetPlayerIDs())

	}
	return
}

// 添加一个PlayerID到一个格子中
func (m *AOIManager) AddPidToGrid(pID, gID int) {
	m.grids[gID].Add(pID)
}

// 移除一个格子中的PlayerID
func (m *AOIManager) RemovePidfromGrid(pID, gID int) {
	m.grids[gID].Remove(pID)
}

// 通过GID获取当前格子的全部playerID
func (m *AOIManager) GetPidsByGid(gID int) (playerIDs []int) {
	playerIDs = m.grids[gID].GetPlayerIDs()
	return
}

// 通过横纵坐标添加一个Player到一个格子中
func (m *AOIManager) AddtoGridByPos(pID int, x, y float32) {
	gID := m.GetGidbyPos(x, y)

	grid := m.grids[gID]
	grid.Add(pID)

}

// 通过横纵坐标把一个Player从对应的格子中删除
func (m *AOIManager) RemoveFromGridByPos(pID int, x, y float32) {
	gID := m.GetGidbyPos(x, y)
	grid := m.grids[gID]
	grid.Remove(pID)
}
