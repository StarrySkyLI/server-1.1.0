package znet

import (
	"fmt"
	"server-1.1.0/network/utils"
	"server-1.1.0/network/ziface"
	"strconv"
)

// 消息处理模块的实现
type MsgHandle struct {
	//存放每个MsgID所对应的处理方法
	Apis map[uint32]ziface.IRouter
	//负责Worker读取任务的消息队列
	TaskQueue []chan ziface.IRequest

	//业务工作Worker池的数量
	WorkerPoolSize uint32
}

// 初始化MsgHandle方法
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize, //从全局配置中获取
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

// 调度、执行对应的Router消息处理方法
func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	// 1 从Request中找到msgID
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgID = ", request.GetMsgID(), "is NOT FOUND! NEED Register")
		return
	}

	//2根据MsgId调度对应Router业务即可
	handler.PreHandlle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgID uint32, router ziface.IRouter) {
	//1 判断 当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgID]; ok {
		//ID已经注册		panic是字符串拼接用strconv.Itoa()来
		panic("repeat api,msgID=" + strconv.Itoa(int(msgID)))
	}
	//2 添加msg和API的绑定关系
	mh.Apis[msgID] = router
	fmt.Println("Add api MsgID =", msgID, "succ!")

}

// 启动一个Worker工作池(开启工作池的动作只能发生一次，一个Zinx框架只能有一个工作池
func (mh *MsgHandle) StartWorkerPool() {
	//根据workerpoolsize分别开启worker，每个worker用一个go来承载
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//当前的worker被启动
		//1 当前的worker对应的channel消息队列 开辟空间第0个worker就用第0个channel。。。
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//2 启动StartOneWorker，阻塞等待消息从channel传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])

	}

}

// 启动一个Worker工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("WorkerID=", workerID, "is started...")
	//不断的阻塞等待对应消息队列的消息
	for {
		select {
		//如果有消息过来，从列的就是一个客户端的Request，执行当前Request所绑定的业务
		case request := <-taskQueue:
			mh.DoMsgHandler(request)

		}
	}

}

// 将消息交给TaskQueue,由Worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	//将消息平均分配给不通过的worker
	//根据客户端建立的ConnID来进行分配
	//基本的平均分配轮询法则
	workerID := request.GetConnection().GetConnId() % mh.WorkerPoolSize
	fmt.Println("Add ConnID =", request.GetConnection().GetConnId(), "request MsgID =", request.GetMsgID(), "to WorkID =", workerID)
	//将消息发送给对应的worker的TaskQueue即可
	mh.TaskQueue[workerID] <- request

}
