package csvs

import "server-1.1.0/utils"

type ConfigHomeItem struct {
	HomeItemId int `json:"HomeItemId"`
	Type       int `json:"Type"`
}

var (
	ConfigHomeItemMap map[int]*ConfigHomeItem
)

func init() {
	ConfigHomeItemMap = make(map[int]*ConfigHomeItem)
	utils.GetCsvUtilMgr().LoadCsv("Home", &ConfigHomeItemMap)
	return
}

func GetHomeItemConfig(HomeItemId int) *ConfigHomeItem {
	return ConfigHomeItemMap[HomeItemId]
}
