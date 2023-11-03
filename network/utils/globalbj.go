package utils

import (
	"encoding/json"
	"os"
	"server-1.1.0/network/ziface"
)

/*
存储一切有关Zinx框架的全局参数，供其他模块使用
一些参数也可以通过 用户根据 zinx.json来配置
*/
type DBConfig struct {
	DBUser     string `json:"dbuser" `
	DBPassword string `json:"dbpassword" `
}
type GlobalObj struct {
	//server
	TcpServer ziface.IServer //当前Zinx的全局Server对象
	Host      string         //当前服务器主机IP
	TcpPort   int            //当前服务器主机监听端口号
	Name      string         //当前服务器名称
	//zinx
	Version          string //当前Zinx版本号
	MaxPacketSize    uint32 //都需数据包的最大值
	MaxConn          int    //当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint32 //当前业务工作Worker池的Goroutine数量
	MaxWorkerTaskLen uint32
	LocalSavePath    string    `json:"localsavepath"` //! 本地存储路径
	DBConfig         *DBConfig `json:"database" `
}

/*
定义一个全局的对象
*/
var GlobalObject *GlobalObj

func (g *GlobalObj) Reload() {
	data, err := os.ReadFile("./conf/zinx.json")
	if err != nil {
		panic(err)
	}
	//将json解析到struct中
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

// 提供一个init方法，初始化当前对象GlobalObject
func init() {
	//如果配置文件没有加载，默认值
	GlobalObject = &GlobalObj{
		Name:             "ZinxServerApp",
		Version:          "v1.0",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          1000,
		MaxPacketSize:    4096,
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024, //每个worker对应的消息队列的任务最大值
	}
	//从conf/zinx.txt 配置文件中加载一些用户配置的参数
	GlobalObject.Reload()
}
