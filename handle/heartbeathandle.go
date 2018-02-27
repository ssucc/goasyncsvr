package handle

import (
	"github.com/astaxie/beego/logs"
	"github.com/golang/protobuf/proto"
	. "github.com/ssucc/goasyncsvr/dispatcherinterface"
	"github.com/ssucc/goasyncsvr/netinterface"
	. "github.com/ssucc/goasyncsvr/packet"
	. "github.com/ssucc/goasyncsvr/protocol"
	"runtime"
	"time"
)

type HeartBeatHandler struct {
	BaseHandler
	reqBody *HeartBeatReq
}

func (h *HeartBeatHandler) Init(packet *Packet, dispatcher IDispatcher, conn netinterface.IConnector) error {
	err := h.BaseHandler.Init(packet, dispatcher, conn)
	if err != nil {
		return err
	}

	h.reqBody = new(HeartBeatReq)
	err = proto.Unmarshal(h.reqMsg.Body, h.reqBody)
	if err != nil {
		logs.Error("unmarshal body failed.err:%s", err)
	} else {
		logs.Info("reqbody:%s", h.reqBody.String())
	}

	return err
}

func (b *HeartBeatHandler) DoWork() error {
	defer func() {
		if err := recover(); nil != err {
			tmpbuf := make([]byte, 0, 1024*1024)
			len := runtime.Stack(tmpbuf, false)
			logs.Error("dowork:%s err:%s len:%d stack:%s", b.reqMsg.String(), err, len, tmpbuf)
		}
	}()

	rspBody := new(HeartBeatRsp)
	rspBody.Remotetime = uint32(time.Now().Unix())
	logs.Info("rspbody:%s", rspBody.String())
	bodybuf, err := proto.Marshal(rspBody)
	if err != nil {
		b.DoError(ERR_TIMEOUT)
		return err
	}

	b.rspBodyBuf = bodybuf
	b.DoRsp()
	return nil
}
