package dispatcherinterface

import (
	"github.com/fsvr/base"
	"github.com/fsvr/netinterface"
	"github.com/fsvr/packet"
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
