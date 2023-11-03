package ziface

// 封包拆包的模块，直接面向TCP链接中的数据流，用于处理TCP粘包问题
type IDataPack interface {
	GetHeadLen() uint32                //获取包头长度方法
	Pack(msg IMessage) ([]byte, error) //封包方法
	Unpack([]byte) (IMessage, error)   //拆包方法
}
