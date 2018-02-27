package dispatcherinterface

import (
	"github.com/ssucc/goasyncsvr/base"
	"github.com/ssucc/goasyncsvr/netinterface"
	"github.com/ssucc/goasyncsvr/packet"
)

type IDispatcher interface {
	GetConf(pid int32) *base.Conf
	DispatchTimeTout(timeout int32)
	Dispatch()
	DispatchReqMsg(pkg *packet.Packet, c netinterface.IConnector)
	Init(conffile string) error
	RegisterCmdHandle(cmd int32, handleName string, regStruct map[string]interface{}) error
	ReloadConf()
	GetMemberByName(name string) interface{}
}
