package ihandle

import (
	. "github.com/ssucc/goasyncsvr/dispatcherinterface"
	"github.com/ssucc/goasyncsvr/netinterface"
	. "github.com/ssucc/goasyncsvr/packet"
)

type InHandler interface {
	//DecodeMsg(packet *Packet) (*Msg, error)
	Init(packet *Packet, dispatcher IDispatcher, conn netinterface.IConnector) error
	DoWork() error
	DoRsp()
	DoError(retCode int32)
	GenRspMsg(rspBody []byte) []byte
}
