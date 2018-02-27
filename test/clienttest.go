package main

import (
	"errors"
	//"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/golang/protobuf/proto"
	"github.com/ssucc/goasyncsvr/network"
	"github.com/ssucc/goasyncsvr/packet"
	. "github.com/ssucc/goasyncsvr/protocol"
	"io"
	"net"
	//"runtime"
	"time"
)

func ReadToBuf(conn *net.Conn, buf []byte) (int, error) {
	if len(buf) == 0 {
		return -1, errors.New("Buf length is zero.!!!")
	}

	l := 0
	for {
		tmp := buf[l:]
		//返回 len = 0， 并且err == syscall.EINVA 连接无效,
		//返回
		recvlen, err := (*conn).Read(tmp)
		if recvlen > 0 {
			logs.Info("recvlen:%d err:%s", recvlen, err)
		}

		if err != nil {
			return -1, err
		}

		if recvlen > 0 {
			//err == nil or err == EOF
			l += recvlen
		}

		if recvlen == 0 && err == io.EOF {
			//连接关闭，没有数据
			return l, nil
		}

		if err != nil {
			//发生错误
			logs.Error("readlen:%d err:%s", l, err)
			return 0, err
		}

		if l > packet.PKG_HEADER_LEN {
			//len > 0, err == nil or err == EOF or len == 0 and err == nil
			//可能被中断，或者数据处理完,先处理一部分数据
			return l, nil
		}
	}
}

func DoOneHeartBeat(beginchan chan bool) {
	<-beginchan
	connector := network.NewConnectorWithHost(":9100", time.Duration(0))
	if connector == nil {
		logs.Error("NewConnector failed.")
		return
	}

	go connector.ReadMsg()
	go connector.WriteMsg()
	reqbody := new(HeartBeatReq)
	reqbody.Ip = connector.Conn.LocalAddr().String()
	bodybyte, _ := proto.Marshal(reqbody)
	reqmsg := new(Msg)
	reqmsg.Head = new(Head)
	reqmsg.Head.Bid = 1000
	reqmsg.Head.Pid = 100001
	reqmsg.Head.Cmd = int32(CMD_HEARTBEAT_REQ)

	reqmsg.Body = bodybyte

	msgbyte, _ := proto.Marshal(reqmsg)
	var headkey [16]byte
	pkg := packet.NewPkg(int32(PID_ID_DATE_1), int32(CMD_HEARTBEAT_REQ), 1, headkey, msgbyte)
	connector.WriteToChan(pkg.Serialize())
	buf, err := connector.ReadFromChan()
	if err != nil {
		logs.Error("read failed.%s", err)
	}

	rsppkg := packet.DecodeToPkg(buf)

	rspmsg := new(Msg)
	proto.Unmarshal(rsppkg.Body, rspmsg)
	logs.Info("rspmsg:%s", rspmsg.String())

	rspbody := new(HeartBeatRsp)
	proto.Unmarshal(rspmsg.Body, rspbody)
	logs.Info("rspbody:%s", rspbody.String())

}

func main() {

	c := make(chan bool)
	for i := 0; i < 10; i++ {
		go DoOneHeartBeat(c)
	}

	close(c)
	select {}
	/*addr, err := net.ResolveTCPAddr("tcp4", "192.168.0.106:9100")
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		fmt.Printf("connect failed. err:%s\n", err)
		return
	}*/
	/*
		conn, err := net.Dial("tcp", "192.168.0.106:9100")
		fmt.Printf("local addr:%s remote addr:%s\n", conn.LocalAddr().String(), conn.RemoteAddr().String())

		reqbody := new(HeartBeatReq)
		reqbody.Ip = conn.LocalAddr().String()
		bodybyte, err := proto.Marshal(reqbody)
		reqmsg := new(Msg)
		reqmsg.Head = new(Head)
		reqmsg.Head.Bid = 1000
		reqmsg.Head.Pid = 100001
		reqmsg.Head.Cmd = int32(CMD_HEARTBEAT_REQ)

		reqmsg.Body = bodybyte

		msgbyte, _ := proto.Marshal(reqmsg)
		var headkey [16]byte
		pkg := packet.NewPkg(int32(PID_ID_DATE_1), int32(CMD_HEARTBEAT_REQ), 1, headkey, msgbyte)
		fmt.Printf("pkglen len:%d pkghead:%s msglen:%d\n", pkg.Header.PkgLen, pkg.String(), len(msgbyte))
		len, err := conn.Write(pkg.Serialize())
		fmt.Printf("write len:%d %s\n", len, err)

		readbuf := make([]byte, 1024)
		readlen, err := ReadToBuf(&conn, readbuf)
		fmt.Printf("status:%d err:%s\n", readlen, err)

		rsppkg := packet.DecodeToPkg(readbuf)

		rspmsg := new(Msg)
		proto.Unmarshal(rsppkg.Body, rspmsg)
		logs.Info("rspmsg:%s", rspmsg.String())

		rspbody := new(HeartBeatRsp)
		proto.Unmarshal(rspmsg.Body, rspbody)
		logs.Info("rspbody:%s", rspbody.String())*/

	return
}
