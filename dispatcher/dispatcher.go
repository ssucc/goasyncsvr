package dispatcher

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/fsvr/base"
	"github.com/fsvr/dispatcherinterface"
	"github.com/fsvr/netinterface"
	"github.com/fsvr/packet"
	"github.com/fsvr/queue"
	"github.com/fsvr/route"
	"reflect"
	"time"
)

type Dispatcher struct {
	dispatcherinterface.IDispatcher
	QueueMng    *queue.QueueManager
	RouteMng    *route.RouteManager
	isStop      bool
	ConfFile    string
	Seq         uint32
	isInit      bool
	CmdToHandle map[int32]reflect.Type
	Conf        *base.Conf
}

func (d *Dispatcher) GetMemberByName(name string) interface{} {
	if name == "Conf" {
		return d.Conf
	}

	return nil
}

func (d *Dispatcher) GetConf(pid int32) *base.Conf {
	return d.Conf
}

func (d *Dispatcher) ReloadConf() {
	conf, err := base.NewConf("xml", d.ConfFile)
	if err != nil {
		logs.Error("New Conf failed.err:%s", err)
		return
	}

	d.Conf = conf

	return
}

func NewDispatcher(conffile string) *dispatcherinterface.IDispatcher {
	var idispatcher dispatcherinterface.IDispatcher
	idispatcher = new(Dispatcher)
	err := idispatcher.Init(conffile)
	if err == nil {
		return &idispatcher
	} else {
		return nil
	}
}

func (d *Dispatcher) RegisterCmdHandle(cmd int32, handleName string, regStruct map[string]interface{}) error {
	if regStruct[handleName] != nil {
		t := reflect.ValueOf(regStruct[handleName]).Type()
		d.CmdToHandle[cmd] = t
		return nil
	}

	return errors.New(fmt.Sprintf("can't find the handlename:%s", handleName))
}

func (d *Dispatcher) Init(conffile string) error {
	qmng := queue.NewQueueManager()
	d.QueueMng = qmng
	d.ConfFile = conffile
	d.isStop = false
	d.CmdToHandle = make(map[int32]reflect.Type)
	return nil
}

func (d *Dispatcher) Stop() {
	d.isStop = true
}

func (d *Dispatcher) Dispatch() {
	var lastCheckStamp int64
	lastCheckStamp = 0
	tc := time.Tick(time.Millisecond * 10)
	for !d.isStop {
		<-tc
		curTime := time.Now().Unix()
		if curTime-lastCheckStamp > 10 {
			d.DispatchTimeTout(10)
			lastCheckStamp = curTime
		}
	}

	logs.Error("Dispatcher End!")
}

func (d *Dispatcher) DispatchTimeTout(timeout int32) {

}

//优化措施，根据cmd自动找到处理的obj，采用预先注册的方式，使用字符串反射构建对应的handle处理对应的cmd
func (d *Dispatcher) DispatchReqMsg(pkg *packet.Packet, c netinterface.IConnector) {
	defer func() {
		if err := recover(); nil != err {
			logs.Error("err:%s", err)
		}
	}()

	//使用反射进行处理
	if d.CmdToHandle[int32(pkg.Header.CMD)] != nil {
		h := reflect.New(d.CmdToHandle[int32(pkg.Header.CMD)])
		f := h.MethodByName("Init")
		if f.IsValid() {
			f.Call([]reflect.Value{reflect.ValueOf(pkg), reflect.ValueOf(d), reflect.ValueOf(c)})
		}

		f1 := h.MethodByName("DoWork")
		if f1.IsValid() {
			f1.Call(make([]reflect.Value, 0))
		}
	} else {
		logs.Error("can't find handler.cmd:%d", int(pkg.Header.CMD))
	}

	/*
			var ihandler ihandle.InHandler
			ihandler = nil

			switch pkg.Header.CMD {
			case int32(CMD_HEARTBEAT_REQ):
				ihandler = new(handle.HeartBeatHandler)
			default:
				logs.Error("Unknown cmd:%d", pkg.Header.CMD)
			}


		if ihandler != nil {
			err := ihandler.Init(pkg, d, c)
			if err == nil {
				go ihandler.DoWork()
			} else {
				logs.Error("Init failed.cmd:%d", pkg.Header.CMD)
			}
		}
	*/
}
