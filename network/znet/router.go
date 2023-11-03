package znet

import "server-1.1.0/network/ziface"

// 实现router时，先嵌入这个baserouter，然后根据需要对这个基类进行重写就好了
type BaseRouter struct {
}

// 这里之所以baseRouter的方法都为空
// 是因为有的Router不希望有PreHandle，posthandle这两个业务
// 所以Router全部继承BaseRouter的好处就是，不需要实现PreHandle，posthandle
// 处理conn业务之前的钩子方法hook
func (br *BaseRouter) PreHandlle(request ziface.IRequest) {}

// 处理conn业务的主方法hook
func (br *BaseRouter) Handle(request ziface.IRequest) {}

// 处理conn业务之后的钩子方法hook
func (br *BaseRouter) PostHandle(request ziface.IRequest) {}
