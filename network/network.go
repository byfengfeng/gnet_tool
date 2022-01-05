package network

import (
	"github.com/Byfengfeng/gnet_tool/code_tool"
	"github.com/Byfengfeng/gnet_tool/inter"
	"github.com/Byfengfeng/gnet_tool/log"
	"github.com/Byfengfeng/gnet_tool/utils"
	"go.uber.org/zap"
	"net"
	"sync"
	"sync/atomic"
)

type NetWork struct {
	*net.TCPConn
	ReadChan  chan []byte
	WriteChan chan []byte
	IsClose   bool
	CloseLock sync.RWMutex
	Ctx       *code_tool.IRequestCtx
	ringByte  *utils.Bytes
}

var (
	count uint32
)

func NewNetWork(c *net.TCPConn) {
	address := c.RemoteAddr().String()
	t := &NetWork{c,
		make(chan []byte),
		make(chan []byte),
		false,
		sync.RWMutex{},
		code_tool.NewIRequestCtx(0, address),
		utils.NewBytes(1024).ReadBytes(),
	}
	code_tool.NewChannel(t)
	t.Start()
}

func (n *NetWork) readBuff()  {
	for {
		newBytes := make([]byte, 1024)
		readLen, err := n.TCPConn.Read(newBytes)
		if err != nil {
			log.Logger.Info(err.Error())
			code_tool.OffLine(n.Ctx.Addr, n.Ctx.Cid)
			return
		}
		if readLen == 0 {
			log.Logger.Info("read off")
			code_tool.OffLine(n.Ctx.Addr, n.Ctx.Cid)
			return
		} else {
			n.ringByte.WriteBytes(uint16(readLen),newBytes[0:readLen-1])
		}
	}
}
func (n *NetWork) read() {
	for {
		if !n.IsClose {
			log.Logger.Info("read off")
			return
		}
		reqBytes := <-n.ringByte.Read()
		if len(reqBytes) == 0 {
			log.Logger.Info("read off")
			return
		} else {
			//读取数据
			code, data := utils.Decode(reqBytes)
			go code_tool.Request(n.Ctx.Addr, n, code, data)
		}
	}
}

func (n *NetWork) write() {
	for {
		if !n.IsClose {
			log.Logger.Info("write off")
			return
		}
		data := <-n.WriteChan
		if len(data) > 0 {
			_, err := n.Write(data)
			if err != nil {
				log.Logger.Error("发送消息异常", zap.Any("err", err))
				return
			}
		} else {
			log.Logger.Info("write off")
			return
		}
	}
}

func (n *NetWork) Start() {
	go n.readBuff()
	go n.read()
	go n.write()
}

func (n *NetWork) GetCtx() interface{} {
	return n.Ctx
}

func (n *NetWork) WriteReadChan(data []byte) {
	n.ReadChan <- data
}

func (n *NetWork) WriteWriteChan(data []byte) {
	n.WriteChan <- data
}

func (n *NetWork) SetIsClose() {
	n.CloseLock.Lock()
	defer n.CloseLock.Unlock()
	if n.IsClose {
		n.IsClose = false
		n.TCPConn.Close()
		n.ringByte.Close()
		close(n.ReadChan)
		close(n.WriteChan)
		log.Logger.Info("close network")
		atomic.AddUint32(&count, 1)
	}
}

func (n *NetWork) CloseCid() {
	code_tool.OffLine(n.Ctx.Addr, n.Ctx.Cid)
}

func GetNetWork(address string) inter.INetwork {
	return code_tool.GetNetWorkByAddr(address)
}

func (n *NetWork) GetNetWorkBy(address string) inter.INetwork {
	return GetNetWork(address)
}

func (n *NetWork) Action(action func()) {
	n.CloseLock.RLock()
	if n.IsClose {
		action()
	}
	n.CloseLock.RUnlock()
}

func (n *NetWork) GetAddr() string {
	return n.RemoteAddr().String()
}

func (n *NetWork) SetCid(cid int64) {
	n.Ctx.Cid = cid
}

func (n *NetWork) SetUid(uid int64) {
	n.Ctx.Uid = uid
}

func GetCloseCount() uint32 {
	return count
}

func SetCount() {
	count = 0
}
