package server

import (
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/ssucc/goasyncsvr/base"
	"github.com/ssucc/goasyncsvr/dispatcherinterface"
	"github.com/ssucc/goasyncsvr/network"
	"net"
)

type FServer struct {
	listenaddr  net.TCPAddr
	isStop      bool
	conf        *base.Conf
	listener    *net.TCPListener
	iDispatcher *dispatcherinterface.IDispatcher
	isInit      bool
	confFile    string
}

func (s *FServer) Init(conffile string, iDispatcher *dispatcherinterface.IDispatcher) error {

	defer func() {
		if err := recover(); nil != err {
			logs.Error("Handle FServer Init recvoer failed. err:%s", err)
		}
	}()

	s.confFile = conffile
	rerr := s.loadconf(s.confFile)
	if rerr != nil {
		logs.Error("load conf failed.err:%s", rerr)
		return rerr
	}

	/*
		conf, err := base.NewConf("xml", conffile)
		if err != nil {
			logs.Error("New Conf failed.err:%s", err)
			return errors.New("New conf Faield.")
		}

		s.conf = conf

		logs.EnableFuncCallDepth(true)
		logs.SetLogFuncCallDepth(int(s.conf.GetFieldInt("loglevel", 3)))
	*/

	host := s.conf.GetFieldStr("host", "eth0:9000")
	nettype := s.conf.GetFieldStr("nettype", "tcp4")
	logs.Info("host:%s nettypes:%s", host, nettype)
	addr, err := net.ResolveTCPAddr(nettype, host)

	if err != nil {
		logs.Error("Can't resolve TCPAddr")
		return errors.New("Can't resolve TCPAddr")
	}

	logs.Info("host:%s nettypes:%s resolveaddr:%s", host, nettype, *addr)

	s.listenaddr = *addr
	ln, err := net.ListenTCP(nettype, addr)
	if err != nil {
		logs.Error("Listen failed.")
		return errors.New("Listen failed.")
	}

	s.listener = ln
	s.iDispatcher = iDispatcher
	if s.iDispatcher != nil {
		go (*(s.iDispatcher)).Dispatch()
	}

	s.isInit = true
	return err
}

func (s *FServer) loadconf(filename string) error {
	conf, err := base.NewConf("ini", filename)
	if err != nil {
		logs.Error("New Conf failed.err:%s", err)
		return err
	}

	s.conf = conf

	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(int(s.conf.GetFieldInt("loglevel", 3)))
	if s.iDispatcher != nil {
		(*(s.iDispatcher)).ReloadConf()
	}

	return nil
}

func (s *FServer) Reload() error {
	return s.loadconf(s.confFile)
}

func (s *FServer) Run() {
	go func() {
		if !s.isInit {
			logs.Error("fserver is not inited. server end.")
			return
		}

		for !s.isStop {
			conn, err := s.listener.AcceptTCP()
			if err != nil {
				// handle error
			}

			logs.Info("recv conn from ip:%s\n", conn.RemoteAddr().String())
			s.DealConn(conn)
		}
	}()
}

func (s *FServer) Stop() {
	s.listener.Close()
	s.isStop = true
}

func (s *FServer) DealConn(conn *net.TCPConn) {
	defer func() {
		if err := recover(); nil != err {
			logs.Error("err:%s", err)
		}
	}()

	connobj := network.NewConnector(conn, s.conf, s.iDispatcher)
	connobj.StartSvr()
}
