package main

import (
	"flag"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/fsvr/dispatcher"
	"github.com/fsvr/handle"
	. "github.com/fsvr/protocol"
	"github.com/fsvr/server"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

func main() {
	tfilename := (flag.String("conf", "./conf/fsvr.conf", "-conf=./conf/fsvr.conf"))
	flag.Parse()
	filename := *tfilename
	logs.Info("filename:%s", filename)
	file, err := os.Open(filename)
	if err != nil {
		logs.Error("open conffile failed.filename:%s", filename)
		return
	}

	filecontent := make([]byte, 1024*10)
	file.Read(filecontent)
	logs.Error("filename:%s content:%s", filename, string(filecontent))

	//init dispatcher
	idispatcher := dispatcher.NewDispatcher(filename)
	(*idispatcher).RegisterCmdHandle(int32(CMD_HEARTBEAT_REQ), "HeartBeatHandle", handle.RegHandleStruct)
	fServer := new(server.FServer)
	fServer.Init(filename, idispatcher)
	fServer.Run()

	var s = make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGKILL, syscall.SIGUSR1)
	//是否收到kill的命令
	for {
		cmd := <-s
		if cmd == syscall.SIGKILL {
			break
		} else if cmd == syscall.SIGUSR1 {
			//如果为siguser1则进行dump内存
			unixtime := time.Now().Unix()
			path := fmt.Sprintf("./heapdump-fserver-%d", unixtime)
			f, err := os.Create(path)
			if nil != err {
				continue
			} else {
				debug.WriteHeapDump(f.Fd())
			}
		}
	}

	fServer.Stop()
	logs.Error("FServer IS STOPPED!")
	return
}
