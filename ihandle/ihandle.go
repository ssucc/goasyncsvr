package ihandle

import (
	. "github.com/fsvr/dispatcherinterface"
	"github.com/fsvr/netinterface"
	. "github.com/fsvr/packet"
)

type InHandler interface {
	//DecodeMsg(packet *Packet) (*Msg, error)
	Init(packet *Packet, dispatcher IDispatcher, conn netinterface.IConnector) error
	DoWork() error
	DoRsp()
	DoError(retCode int32)
	GenRspMsg(rspBody []byte) []byte
}
