package network

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/util/bytespool"
	"net"
	"sync"
	"time"
)

const (
	Default_ReadDeadline    = time.Second * 30 //默认读超时30s
	Default_WriteDeadline   = time.Second * 30 //默认写超时30s
	Default_MaxConnNum      = 1000000          //默认最大连接数
	Default_PendingWriteNum = 100000           //单连接写消息Channel容量
	Default_MinMsgLen       = 2                //最小消息长度2byte
	Default_LenMsgLen       = 2                //包头字段长度占用2byte
	Default_MaxMsgLen       = 65535            //最大消息长度
)

type TCPServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	ReadDeadline    time.Duration
	WriteDeadline   time.Duration

	NewAgent   func(conn Conn) Agent
	ln         net.Listener
	conns      ConnSet
	mutexConns sync.Mutex
	wgLn       sync.WaitGroup
	wgConns    sync.WaitGroup

	MsgParser
}

func (server *TCPServer) Start() error {
	err := server.init()
	if err != nil {
		return err
	}

	server.wgLn.Add(1)
	go server.run()

	return nil
}

func (server *TCPServer) init() error {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("listen tcp fail,error:%s", err.Error())
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = Default_MaxConnNum
		log.Debugf("invalid MaxConnNum,reset:%d", server.MaxConnNum)
	}

	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = Default_PendingWriteNum
		log.Debugf("invalid PendingWriteNum,reset:%d", server.PendingWriteNum)
	}

	if server.LenMsgLen <= 0 {
		server.LenMsgLen = Default_LenMsgLen
		log.Debugf("invalid LenMsgLen,reset:%d", server.LenMsgLen)
	}

	if server.MaxMsgLen <= 0 {
		server.MaxMsgLen = Default_MaxMsgLen
		log.Debugf("invalid MaxMsgLen,reset:%d", server.MaxMsgLen)
	}

	maxMsgLen := server.MsgParser.getMaxMsgLen()
	if server.MaxMsgLen > maxMsgLen {
		server.MaxMsgLen = maxMsgLen
		log.Debugf("invalid MaxMsgLen,reset:%d", maxMsgLen)
	}

	if server.MinMsgLen <= 0 {
		server.MinMsgLen = Default_MinMsgLen
		log.Debugf("invalid MinMsgLen,reset:%d", server.MinMsgLen)
	}

	if server.WriteDeadline == 0 {
		server.WriteDeadline = Default_WriteDeadline
		log.Debugf("invalid WriteDeadline,reset:%d", int64(server.WriteDeadline.Seconds()))
	}

	if server.ReadDeadline == 0 {
		server.ReadDeadline = Default_ReadDeadline
		log.Debugf("invalid ReadDeadline,reset:%d", int64(server.ReadDeadline.Seconds()))
	}

	if server.NewAgent == nil {
		return errors.New("NewAgent must not be nil")
	}

	server.ln = ln
	server.conns = make(ConnSet, 2048)
	server.MsgParser.Init()

	return nil
}

func (server *TCPServer) SetNetMemPool(memPool bytespool.IBytesMemPool) {
	server.IBytesMemPool = memPool
}

func (server *TCPServer) GetNetMemPool() bytespool.IBytesMemPool {
	return server.IBytesMemPool
}

func (server *TCPServer) run() {
	defer server.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				log.Infof("accept fail,error:%s,sleep time:%d", err.Error(), tempDelay)
				tempDelay = min(1*time.Second, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}

		conn.(*net.TCPConn).SetLinger(0)
		conn.(*net.TCPConn).SetNoDelay(true)
		tempDelay = 0

		server.mutexConns.Lock()
		if len(server.conns) >= server.MaxConnNum {
			server.mutexConns.Unlock()
			conn.Close()
			log.Warn("too many connections")
			continue
		}

		server.conns[conn] = struct{}{}
		server.mutexConns.Unlock()
		server.wgConns.Add(1)

		tcpConn := newNetConn(conn, server.PendingWriteNum, &server.MsgParser, server.WriteDeadline)
		agent := server.NewAgent(tcpConn)

		go func() {
			agent.Run()
			// cleanup
			tcpConn.Close()
			server.mutexConns.Lock()
			delete(server.conns, conn)
			server.mutexConns.Unlock()
			agent.OnClose()

			server.wgConns.Done()
		}()
	}
}

func (server *TCPServer) Close() {
	server.ln.Close()
	server.wgLn.Wait()

	server.mutexConns.Lock()
	for conn := range server.conns {
		conn.Close()
	}
	server.conns = nil
	server.mutexConns.Unlock()
	server.wgConns.Wait()
}
