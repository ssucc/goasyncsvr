package handle

import (
	"github.com/astaxie/beego/logs"
	. "github.com/fsvr/dispatcherinterface"
	"github.com/fsvr/ihandle"
	"github.com/fsvr/netinterface"
	. "github.com/fsvr/packet"
	. "github.com/fsvr/protocol"
	"github.com/golang/protobuf/proto"
)

//定义handle的映射，每次增加一个handle，就需要在这里添加映射
var RegHandleStruct map[string]interface{}

func init() {
	RegHandleStruct = make(map[string]interface{})
	RegHandleStruct["HeartBeatHandle"] = HeartBeatHandler{}
}

type BaseHandler struct {
	ihandle.InHandler
	dispatcher IDispatcher
	conn       netinterface.IConnector

	pkg        *Packet
	reqMsg     *Msg
	rspMsg     *Msg
	stopChan   chan bool
	uin        uint64
	retCode    int32
	rspBodyBuf []byte
}

func (b *BaseHandler) Stop() {
	b.stopChan <- true
}

func (b *BaseHandler) Run() {
	for {
		select {
		case <-b.stopChan:
			logs.Error("handle receive stop singal")
		default:
			b.DoWork()
		}
	}
}

func (b *BaseHandler) Init(packet *Packet, dispatcher IDispatcher, conn netinterface.IConnector) error {
	b.dispatcher = dispatcher
	b.conn = conn
	b.pkg = packet
	b.stopChan = make(chan bool)
	b.rspBodyBuf = make([]byte, 0, 1024*1024)

	b.reqMsg = new(Msg)
	err := proto.Unmarshal(b.pkg.Body, b.reqMsg)
	if err != nil {
		logs.Error("can't parse pkt.PkgLen:%d ProductID:%d CMD:%d CliSeq:%d", b.pkg.Header.PkgLen, b.pkg.Header.ProductID, b.pkg.Header.CMD, b.pkg.Header.CliSeq)
	} else {
		logs.Info("msg:%s", b.reqMsg.String())
	}

	// reqbody := new(HeartBeatReq)
	// err = proto.Unmarshal(b.reqMsg.Body, reqbody)
	// if err != nil {
	// 	logs.Error("unmarshal body failed.err:%s", err)
	// } else {
	// 	logs.Info("reqbody:%s", reqbody.String())
	// }

	return err
}

func (b *BaseHandler) DoWork() error {
	logs.Error("not implement at this scope")
	return nil
}

func (b *BaseHandler) DoRsp() {
	rspbuf := b.GenRspMsg(b.rspBodyBuf)
	rspPkg := NewRspPkg(b.pkg.Header.ProductID, b.pkg.Header.CMD, b.pkg.Header.CliSeq, b.pkg.Header.HeadKey, rspbuf)
	rspbyte := rspPkg.Serialize()
	err := b.conn.WriteToChan(rspbyte)
	if err != nil {
		logs.Error("rsp failed.err:%s", err)
	}
}

func (b *BaseHandler) DoError(retCode int32) {
	b.retCode = retCode
	b.DoRsp()
	return
}

func (b *BaseHandler) GenRspMsg(rspBody []byte) []byte {

	b.rspMsg = new(Msg)
	b.rspMsg.Head = new(Head)
	*(b.rspMsg.Head) = *(b.reqMsg.Head)
	b.rspMsg.Head.Retcode = b.retCode
	b.rspMsg.Head.Cmd = b.reqMsg.Head.Cmd + int32(CMD_REQ_RSP_SPAN)
	b.rspMsg.Body = rspBody
	logs.Info("rspMsg:%s len:%d", b.rspMsg.String(), len(rspBody))
	rspbuf, err := proto.Marshal(b.rspMsg)

	if err != nil {
		logs.Error("uin:%u Marshal failed.ret:%d", b.uin, b.retCode)
		return nil
	}

	return rspbuf
}
