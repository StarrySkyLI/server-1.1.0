package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"server-1.1.0/network/utils"
	"server-1.1.0/network/ziface"
)

// 封包，拆包的具体模块
type DataPack struct {
}

// 拆包封包的实例的一个初始化方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

// 获取包头长度方法
func (dp *DataPack) GetHeadLen() uint32 {
	//datalen uint32(4字节）+ID uint32（4字节）
	return 8

}

// 封包方法   | datalen | msgID | data |
func (dp *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	//创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})
	//将datalen写入databuff中  binary.Write二进制写入
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}
	//将MsgId写入databuff中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}
	//将data数据写入databuff中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil

}

// 拆包方法（将包的Head信息读出来）之后再根据head信息里面的data长度，再一次读
func (dp *DataPack) Unpack(binaryData []byte) (ziface.IMessage, error) {
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewBuffer(binaryData)

	//只解压head信息，得到datalen和MsgID
	msg := &Message{}

	//读datalen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}
	//读MsgId
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}
	//判断datalen是否超过我们允许的最大包长度
	if utils.GlobalObject.MaxPacketSize > 0 && msg.DataLen > utils.GlobalObject.MaxPacketSize {
		return nil, errors.New("too Large msg data recv!")
	}
	return msg, nil

}
