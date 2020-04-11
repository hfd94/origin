package rpc

import (
	"fmt"
	"github.com/duanhf2012/origin/log"
	"github.com/duanhf2012/origin/network"
	"math"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Client struct {
	blocalhost bool
	network.TCPClient
	conn *network.TCPConn

	pendingLock sync.RWMutex
	startSeq uint64
	pending map[uint64]*Call
}

func (slf *Client) NewClientAgent(conn *network.TCPConn) network.Agent {
	slf.conn = conn
	slf.pending = map[uint64]*Call{}
	return slf
}

func (slf *Client) Connect(addr string) error {
	slf.Addr = addr
	if strings.Index(addr,"localhost") == 0 {
		slf.blocalhost = true
		return nil
	}
	slf.ConnNum = 1
	slf.ConnectInterval = time.Second*2
	slf.PendingWriteNum = 10000
	slf.AutoReconnect = true
	slf.LenMsgLen =2
	slf.MinMsgLen = 2
	slf.MaxMsgLen = math.MaxUint16
	slf.NewAgent =slf.NewClientAgent
	slf.LittleEndian = LittleEndian

	slf.pendingLock.Lock()
	for _,v := range slf.pending {
		v.Err = fmt.Errorf("node is disconnect.")
		v.done <- v
	}
	slf.pending = map[uint64]*Call{}
	slf.pendingLock.Unlock()
	slf.Start()
	return nil
}

func (slf *Client) AsycGo(rpcHandler IRpcHandler,serviceMethod string,callback reflect.Value, args interface{},replyParam interface{}) error {
	call := new(Call)
	call.Reply = replyParam
	call.callback = &callback
	call.rpcHandler = rpcHandler
	call.ServiceMethod = serviceMethod

	request := &RpcRequest{}
	request.NoReply = false
	call.Arg = args
	slf.pendingLock.Lock()
	slf.startSeq += 1
	call.Seq = slf.startSeq
	request.Seq = slf.startSeq
	slf.pending[call.Seq] = call
	slf.pendingLock.Unlock()
	request.ServiceMethod = serviceMethod
	var herr error
	request.InParam,herr = processor.Marshal(args)
	if herr != nil {
		return herr
	}

	bytes,err := processor.Marshal(request)
	if err != nil {
		return err
	}
	if slf.conn == nil {
		return fmt.Errorf("Rpc server is disconnect,call %s is fail!",serviceMethod)
	}
	err = slf.conn.WriteMsg(bytes)
	if err != nil {
		call.Err = err
	}

	return call.Err
}

func (slf *Client) Go(noReply bool,serviceMethod string, args interface{},reply interface{}) *Call {
	call := new(Call)
	call.done = make(chan *Call,1)
	call.Reply = reply
	call.ServiceMethod = serviceMethod
	request := &RpcRequest{}
	request.NoReply = noReply
	call.Arg = args
	slf.pendingLock.Lock()
	slf.startSeq += 1
	call.Seq = slf.startSeq
	request.Seq = slf.startSeq
	slf.pending[call.Seq] = call
	slf.pendingLock.Unlock()

	request.ServiceMethod = serviceMethod
	var herr error
	request.InParam,herr = processor.Marshal(args)
	if herr != nil {
		call.Err = herr
		return call
	}

	bytes,err := processor.Marshal(request)
	if err != nil {
		call.Err = err
		return call
	}

	if slf.conn == nil {
		call.Err = fmt.Errorf("call %s is fail,rpc client is disconnect.",serviceMethod)
		return call
	}
	
	err = slf.conn.WriteMsg(bytes)
	if err != nil {
		call.Err = err
	}

	return call
}

type RequestHandler func(Returns interface{},Err *RpcError)

type RpcRequest struct {
	//packhead
	Seq uint64             // sequence number chosen by client
	ServiceMethod string   // format: "Service.Method"
	NoReply bool           //是否需要返回
	//packbody
	InParam []byte

	//other data
	localReply interface{}
	localParam interface{} //本地调用的参数列表
	requestHandle RequestHandler
	callback *reflect.Value
}

type RpcResponse struct {
	//head
	Seq           uint64   // sequence number chosen by client
	Err *RpcError

	//returns
	Reply []byte
}


func (slf *Client) Run(){
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			err := fmt.Errorf("%v: %s\n", r, buf[:l])
			log.Error("core dump info:%+v",err)
		}
	}()

	for {
		bytes,err := slf.conn.ReadMsg()
		if err != nil {
			log.Error("rpcClient %s ReadMsg error:%+v",slf.Addr,err)
			return
		}
		//1.解析head
		respone := &RpcResponse{}
		err = processor.Unmarshal(bytes,respone)
		if err != nil {
			log.Error("rpcClient Unmarshal head error,error:%+v",err)
			continue
		}

		slf.pendingLock.Lock()
		v,ok := slf.pending[respone.Seq]
		if ok == false {
			log.Error("rpcClient cannot find seq %d in pending",respone.Seq)
			slf.pendingLock.Unlock()
		}else  {
			delete(slf.pending,respone.Seq)
			slf.pendingLock.Unlock()
			v.Err = nil

			if len(respone.Reply) >0 {
				err = processor.Unmarshal(respone.Reply,v.Reply)
				if err != nil {
					log.Error("rpcClient Unmarshal body error,error:%+v",err)
					v.Err = err
				}
			}

			if respone.Err != nil {
				v.Err= respone.Err
			}

			if v.callback!=nil && v.callback.IsValid() {
				 v.rpcHandler.(*RpcHandler).callResponeCallBack<-v
			}else{
				v.done <- v
			}
		}
	}
}

func (slf *Client) OnClose(){
}

func (slf *Client) IsConnected() bool {
	return slf.conn!=nil && slf.conn.IsConnected()==true
}