package network

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/fsvr/base"
	. "github.com/fsvr/dispatcherinterface"
	"github.com/fsvr/netinterface"
	"github.com/fsvr/packet"
	"io"
	"net"
	"time"
)

type SvrAddr struct {
	host    string
	port    int32
	nettype string
}

type Connector struct {
	netinterface.IConnector
	Conn          *net.TCPConn
	Conf          *base.Conf
	isStop        bool
	CliAddr       string
	WriteChannel  chan []byte
	writeNotify   chan bool
	ReadChannel   chan []byte
	readNotify    chan bool
	PkgDispatcher *IDispatcher
}

/*
	support for conn ?
	create new connector from tcp.conn, base.conf, dispatcher
*/
func NewConnector(newconn *net.TCPConn, newconf *base.Conf, dispatcher *IDispatcher) *Connector {

	return &Connector{
		Conn:          newconn,
		Conf:          newconf,
		isStop:        false,
		CliAddr:       "",
		WriteChannel:  make(chan []byte, newconf.GetFieldInt("MaxNumOfMsgInSendChan", 100000)),
		writeNotify:   make(chan bool),
		ReadChannel:   make(chan []byte, newconf.GetFieldInt("MaxNumOfMsgInRecvChan", 100000)),
		readNotify:    make(chan bool),
		PkgDispatcher: dispatcher,
	}
}

//Client Mode
func NewConnectorWithHost(host string, duration time.Duration) *Connector {
	connector := new(Connector)
	err := connector.Connect(host, duration)
	if err != nil {
		logs.Error("err:%v", err)
		return nil
	}

	connector.WriteChannel = make(chan []byte, 10)
	connector.ReadChannel = make(chan []byte, 10)
	connector.writeNotify = make(chan bool)
	connector.readNotify = make(chan bool)
	connector.isStop = false
	connector.CliAddr = connector.Conn.LocalAddr().String()
	return connector
}

func (c *Connector) StartSvr() {
	defer func() {
		if err := recover(); nil != err {
			logs.Error("handle conn recvoer failed. CliAddr:%s err:%v", c.CliAddr, err)
		}
	}()

	c.isStop = false

	c.Conn.SetKeepAlive(true)
	c.Conn.SetNoDelay(true)
	if c.Conf != nil {
		c.Conn.SetKeepAlivePeriod(time.Duration(c.Conf.GetFieldInt("KeepAlivePeriod", 30)) * time.Second)
		c.Conn.SetReadBuffer(int(c.Conf.GetFieldInt("SockRecvBufLen", 1024*1024*1)))
		c.Conn.SetWriteBuffer(int(c.Conf.GetFieldInt("SockSendBufLen", 1024*1024*1)))
	}

	go c.ReadMsg()
	go c.DispatchMsg()
	go c.WriteMsg()
}

/*发送消息*/
func (c *Connector) WriteToChan(buf []byte) error {
	defer func() {
		if err := recover(); nil != err {
			logs.Error("err:%v", err)
		}
	}()

	if !c.isStop {
		select {
		case c.WriteChannel <- buf:
			if len(c.WriteChannel) == 1 {
				//放进去的是第一个包，通知发送出去
				c.writeNotify <- true
			}

			return nil
		default:
			return errors.New(fmt.Sprintf("WRITE CHANNLE FULL chanlen:%d", len(c.WriteChannel)))
		}
	}

	return errors.New(fmt.Sprintf("Session|CLOSED"))
}

/*接受消息，长连接的客户端使用调用此函数以便获得消息*/
func (c *Connector) ReadFromChanLong() ([]byte, error) {
	defer func() {
		if err := recover(); nil != err {
			logs.Error("err:%v", err)
		}
	}()

	for !c.isStop {
		<-c.readNotify
		select {
		case buf := <-c.ReadChannel:
			if len(c.ReadChannel) > 0 {
				//select防止阻塞
				select {
				case c.readNotify <- true:
				default:
					break
				}
			}

			return buf, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Session|CLOSED"))
}

/*接受消息，仅限于一个单一同步客户端使用*/
func (c *Connector) ReadFromChan() ([]byte, error) {
	defer func() {
		if err := recover(); nil != err {
			logs.Error("err:%v", err)
		}
	}()

	for !c.isStop {
		select {
		case buf := <-c.ReadChannel:
			return buf, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("Session|CLOSED"))
}

func (c *Connector) ReadMsg() {
	defer func() {
		if err := recover(); nil != err {
			logs.Error("handle conn recvoer failed. CliAddr:%s err:%v", c.CliAddr, err)
		}
	}()
	//最大允许的包为8M
	buf := make([]byte, 1024*1024*10)
	var bufidx int = 0
	for !c.isStop {
		readlen, err := c.ReadToBuf(buf[bufidx:])
		logs.Info("idx:%d readlen:%d", bufidx, readlen)
		if readlen > 0 && err == nil {
			//读到> 1个包头的数据, 返回已处理的长度，buf需要移动数据
			totallen := readlen + bufidx
			gotlen, err1 := c.GetMessage(buf, totallen)
			if err1 != nil {
				//或者发生了错误， 发生了错误需要中断吗？
				logs.Error("gotlen:%d err:%v", gotlen, err1)
				c.Close()
				break
			}

			if gotlen > 0 {
				//去掉已处理的数据，将数据在缓存中往前移动
				copy(buf, buf[gotlen:])
			} //gotlen = 0, 不移动

			bufidx = totallen - gotlen //总长度减去已处理的长度
		}

		//发生错误就立即关闭
		if err != nil {
			c.Close()
		}
	}
}

//做服务端需要启动一个go routine运行此函数, 如果是客户端，就不要
func (c *Connector) DispatchMsg() {
	defer func() {
		if err := recover(); nil != err {
			logs.Error("handle conn recvoer failed. CliAddr:%s err:%v", c.CliAddr, err)
		}
	}()

	var laststamp int64 = time.Now().Unix()
	for !c.isStop {
		<-c.readNotify
		//可以一次多读出几个消息，连续发送
		select {
		case msg := <-c.ReadChannel:
			logs.Info("get a msg from channel. len:%d", len(msg))
			c.ProcessMsg(msg)
			if len(c.ReadChannel) > 0 {
				select {
				case c.readNotify <- true:
				default:
					break
				}
			}
		default:
			var curstamp int64 = time.Now().Unix()
			if curstamp-laststamp > 30 {
				logs.Info("no new msg.")
				laststamp = curstamp
			}
		}
	}
}

func (c *Connector) ProcessMsg(buf []byte) {
	defer func() {
		if err := recover(); nil != err {
			logs.Error("%s", err)
		}
	}()
	//完整的一个包
	//len  = 32 + msglen
	pkg := packet.DecodeToPkg(buf)
	logs.Info("pkglen:%d pkgheader:%s", pkg.Header.PkgLen, pkg.String())
	if pkg == nil {
		return
	}

	if c.PkgDispatcher != nil {
		(*(c.PkgDispatcher)).DispatchReqMsg(pkg, c)
	}
}

func (c *Connector) WriteMsg() {
	defer func() {
		if err := recover(); nil != err {
			logs.Error("handle conn recvoer failed. CliAddr:%s err:%v", c.CliAddr, err)
		}
	}()

	var laststamp int64 = time.Now().Unix()
	for !c.isStop {
		<-c.writeNotify
		select {
		case msg := <-c.WriteChannel:
			logs.Info("get a msg from channel. len:%d", len(msg))
			c.write(msg)
			if len(c.WriteChannel) > 0 {
				//如果你们还有数据，继续通知发送
				//如果在发送期间已经有写数据的，如果直接往chan中写通知，那么就造成了死锁，使用select就不会造成死锁,select 会执行default选项
				select {
				case c.writeNotify <- true:
				default:
					break
				}
			}
		default:
			var curstamp int64 = time.Now().Unix()
			if curstamp-laststamp > 30 {
				logs.Info("msg is nil.")
				laststamp = curstamp
			}
		}
	}
}

func (c *Connector) Connect(host string, duration time.Duration) error {
	var conn net.Conn
	var err error
	if duration != time.Duration(0) {
		conn, err = net.DialTimeout("tcp", host, duration*time.Second)
	} else {
		conn, err = net.Dial("tcp", host)
	}

	if err != nil {
		logs.Error("Connect Failed. host:%s, err:%v", host, err)
		return err
	}

	TConn, ok := (conn).(*net.TCPConn)
	if ok {
		c.Conn = TConn
	}

	return nil
}

func (c *Connector) Close() {
	if !c.isStop {
		c.Conn.Close()
		c.isStop = true
		c.WriteChannel = nil
		logs.Error("connector is closed.cliaddr:%s", c.CliAddr)
	}
}

//获取buf的完整的包的数据，并处理
func (c *Connector) GetMessage(buf []byte, buflen int) (int, error) {
	var idx int = 0
	l := buflen
	for l > packet.PKG_HEADER_LEN {
		pkglen := CheckComplete(buf[idx:buflen], buflen)
		logs.Info("pkglen:%d chanlen:%d", pkglen, len(c.ReadChannel))
		if pkglen > 0 {
			//一个完整包数据的长度
			bufnew := make([]byte, pkglen)
			copy(bufnew, buf[idx:(idx+pkglen)])
			c.ReadChannel <- bufnew
			//如果
			logs.Info("pkglen:%d chanlen:%d", pkglen, len(c.ReadChannel))
			if len(c.ReadChannel) == 1 {
				//放进去的是第一个,通知
				c.readNotify <- true
			}

			l -= pkglen
			idx += pkglen
		} else if pkglen < 0 {
			//数据损坏，需要关闭连接
			return -1, nil
		} else if pkglen == 0 {
			//没有完整的包，继续收数据
			break
		}
	}

	return idx, nil
}

/*
type Reader
type Reader interface {
    Read(p []byte) (n int, err error)
}
Reader接口用于包装基本的读取方法.
Read方法读取len(p)字节数据写入p。它返回写入的字节数和遇到的任何错误。
即使Read方法返回值n < len(p)，本方法在被调用时仍可能使用p的全部长度作为暂存空间。
如果有部分可用数据，但不够len(p)字节，Read按惯例会返回可以读取到的数据，而不是等待更多数据。
当Read在读取n > 0个字节后遭遇错误或者到达文件结尾时，会返回读取的字节数。
它可能会在该次调用返回一个非nil的错误，或者在下一次调用时返回0和该错误。
一个常见的例子，Reader接口会在输入流的结尾返回非0的字节数，返回值err == EOF或err == nil。
但不管怎样，下一次Read调用必然返回(0, EOF)。调用者应该总是先处理读取的n > 0字节再处理错误值。
这么做可以正确的处理发生在读取部分数据后的I/O错误，也能正确处理EOF事件。
如果Read的某个实现返回0字节数和nil错误值，表示被阻碍；调用者应该将这种情况视为未进行操作。

type Writer
type Writer interface {
    Write(p []byte) (n int, err error)
}
Writer接口用于包装基本的写入方法。
Write方法len(p) 字节数据从p写入底层的数据流。它会返回写入的字节数(0 <= n <= len(p))和
遇到的任何导致写入提取结束的错误。Write必须返回非nil的错误，如果它返回的 n < len(p)。
Write不能修改切片p中的数据，即使临时修改也不行。
*/

func (c *Connector) ReadToBuf(buf []byte) (int, error) {
	if len(buf) == 0 {
		return -1, errors.New("Buf length is zero.!!!")
	}

	l := 0
	for {
		tmp := buf[l:]
		//返回 len = 0， 并且err == syscall.EINVA 连接无效,
		//返回
		len, err := c.Conn.Read(tmp)
		logs.Info("readlen:%d err:%v", len, err)
		if err != nil {
			return -1, err
		}

		if len > 0 {
			//err == nil or err == EOF
			l += len
		}

		if len == 0 && err == io.EOF {
			//连接关闭，没有数据
			c.Close()
			return l, nil
		}

		if err != nil {
			//发生错误
			logs.Error("readlen:%d err:%v", l, err)
			return 0, err
		}

		if l > packet.PKG_HEADER_LEN {
			//len > 0, err == nil or err == EOF or len == 0 and err == nil
			//可能被中断，或者数据处理完,先处理一部分数据
			return l, nil
		}
	}
}

func (c *Connector) write(msg []byte) {
	msglen := len(msg)
	if msglen < 32 || msglen > 8*1024*1024 {
		logs.Error("msg is too short or long. msglen:%d", msglen)
		return
	}

	writebuf := make([]byte, len(msg))
	//buf := bytes.NewBuffer(writebuf)
	//msgbodylen := len(msg) - 28
	//写入包实际的长度
	//Write(buf, binary.BigEndian, msgbodylen)
	copy(writebuf, msg)
	idx := 0
	leftlen := len(writebuf)
	logs.Info("leftlen:%d", leftlen)
	for leftlen > 0 {
		sendlen, err := c.Conn.Write(writebuf[idx:])
		if err != nil {
			logs.Error("write faield.err:%v", err)
			break
		}

		if sendlen < leftlen {
			idx += sendlen
			leftlen -= sendlen
			logs.Info("totallen:%d sendlen:%d leftlen:%d", msglen, sendlen, leftlen)
			if err != nil {

			}
		} else {
			idx += sendlen
			leftlen -= sendlen
			logs.Info("totallen:%d sendlen:%d leftlen:%d", msglen, sendlen, leftlen)
		}
	}
}
