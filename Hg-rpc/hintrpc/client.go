package hintrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hintrpc/codec"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// 对 net/rpc 而言，一个函数需要能够被远程调用，需要满足五个条件
// 1.方法类型对外暴露
// 2.方法对外暴露
// 3.方法有2个参数，均对外暴露
// 4.方法第2个参数为指针
// 5.方法有1个error返回值
// e.g.
// func (t *T) MethodName(argType T1, replyType *T2) error

// Call represents an active RPC
// 单个RPC请求封装
type Call struct {
	Seq           uint64     // 请求序号，可以理解为id，用于区分不同请求
	ServiceMethod string     // 调用方法 e.g."User.GetUser"
	Args          any        // 请求方法参数
	Reply         any        // 返回参数
	Err           error      // 当error时有值
	Done          chan *Call // call完成后，进入channel
}

// 为了支持异步调用，Call 结构体中添加了一个字段 Done，Done 的类型是 chan *Call，当调用结束时，会调用 call.done() 通知调用方
func (call *Call) done() {
	call.Done <- call
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	cc       codec.Codec      // 解码器，用来序列化将要发送出去的请求，以及反序列化接收到的响应
	opt      *Option          // Server定义(MagicNumber/CodecType)
	sending  sync.Mutex       // 防止信息交织
	header   codec.Header     // 报文头，header 只有在请求发送时才需要，而请求发送是互斥的，因此每个客户端只需要一个，声明在 Client 结构体中可以复用
	mu       sync.Mutex       // 客户端操作防止异步执行
	seq      uint64           // 请求序号，可以理解为id，用于区分不同请求
	pending  map[uint64]*Call // 阻塞等待的RPC请求
	closing  bool             // 用户调用Close方法
	shutdown bool             // server 通知 本客户端停止，一般是有错误发生导致
}

// clientResult 封装Dial后返回的client，方便做超时处理
type clientResult struct {
	client *Client
	err    error
}

// newClientFunc 用于dialTimeout调用时的封装
type newClientFunc func(conn net.Conn, opt *Option) (client *Client, err error)

// dialTimeout 在dial的前提下做超时验证，不超时返回client，超时做对应处理
func dialTimeout(f newClientFunc, network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	// net.Dial -> net.DialTimeout
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	// close the connection if client is nil
	// 获取不到client的情况
	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()
	ch := make(chan clientResult)
	// 使用子协程执行 NewClient，执行完成后则通过信道 ch 发送结果
	go func() {
		client, err := f(conn, opt)
		ch <- clientResult{client: client, err: err}
	}()
	// 不设置连接超时时间，直接返回
	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}
	// 设置超时时间后，根据超时时间进行select
	// 如果 time.After() 信道先接收到消息，则说明 NewClient 执行超时，返回错误。
	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnectTimeout)
	case result := <-ch:
		return result.client, result.err
	}
}

// 保证Client实现了Close方法
var _ io.Closer = (*Client)(nil)

var ErrShuttingDown = errors.New("connection is shutting down")

func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		return ErrShuttingDown
	}
	client.closing = true
	return client.cc.Close()
}

// IsAvailable return true if the client does work
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}

// registerCall 传入一个RPC请求，将请求放入pending中，client赋予call一个独立的seq并返回
// 1. 判断客户端是否可用
// 2. 请求seq放入pending中
// 3. 更新client的seq
// 4. 返回请求的seq或发生的错误，当发生错误时，返回seq为0
// warn. lock support
func (client *Client) registerCall(call *Call) (seq uint64, err error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.shutdown || client.closing {
		return 0, ErrShuttingDown
	}
	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

// removeCall 传入一个seq，返回被移除的call指针
// warn. lock support
func (client *Client) removeCall(seq uint64) (call *Call) {
	client.mu.Lock()
	defer client.mu.Unlock()
	resCall := client.pending[seq]
	delete(client.pending, seq)
	return resCall
}

// terminateCalls 传入一个错误，终止所有的请求
// warn.lock support
func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.shutdown = true
	for _, call := range client.pending {
		call.Err = err
		call.done()
	}
}

// receive 接收响应
// 3种情况：
// call 不存在，可能是请求没有发送完整，或者因为其他原因被取消，但是服务端仍旧处理了
// call 存在，但服务端处理出错，即 h.Error 不为空
// call 存在，服务端处理正常，那么需要从 body 中读取 Reply 的值
func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			// 读取报文头出现错误，直接跳出
			break
		}
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			err = client.cc.ReadBody(nil)
		case h.Error != "":
			call.Err = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Err = errors.New("reading body error " + err.Error())
			}
			call.done()
		}
	}
	// 出现错误时，调用 terminateCalls
	client.terminateCalls(err)
}

// NewClient 传入Option以及connection
func NewClient(conn net.Conn, opt *Option) (client *Client, err error) {
	newCodecF := codec.NewCodecFuncMap[opt.CodecType]
	if newCodecF == nil {
		err = fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client : codec error :", err)
		return nil, err
	}
	// 发送 Option 信息给服务端
	if err = json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client : options error :", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(newCodecF(conn), opt), nil
}

// newClientCodec 根据option创建新的client并开始receive
// NewClient 中已经进行过可行性校验，直接返回Client指针
func newClientCodec(cc codec.Codec, opt *Option) *Client {
	resClient := &Client{
		seq:     1, // seq 由1开始，0代表错误call
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	// 创建一个线程运行receive
	go resClient.receive()
	return resClient
}

// parseOptions 处理用户传入 Option，若未传，返回默认值
func parseOptions(opts ...*Option) (*Option, error) {
	// 未传值
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	// 传入过多opt值
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	// 确认返回值正确
	opt.MagicNumber = MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

// Dial 用户传入服务端地址，创建Client，Option为可选
func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	return dialTimeout(NewClient, network, address, opts...)
}

// send 客户端发送请求
func (client *Client) send(call *Call) {
	// client sending上锁，防止报文交织
	client.sending.Lock()
	defer client.sending.Unlock()
	// register call
	seq, err := client.registerCall(call)
	if err != nil {
		call.Err = err
		call.done()
		return
	}
	// 准备请求报文头
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""
	// 编码&发送请求
	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		if call != nil {
			call.Err = err
			call.done()
		}
	}
}

// Go 是暴露给用户的PRC服务调用接口，返回Call实例
func (client *Client) Go(serviceMethod string, args, reply any, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	client.send(call)
	return call
}

// Call 阻塞call Done，等待响应返回
// 超时处理通过context完成
// 用户可以使用 context.WithTimeout 创建具备超时检测能力的 context 对象来控制
// ctx, _ := context.WithTimeout(context.Background(), time.Second)
func (client *Client) Call(ctx context.Context, serviceMethod string, args, reply any) error {
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 10))
	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		return errors.New("rpc client: call failed: " + ctx.Err().Error())
	case call := <-call.Done:
		return call.Err
	}
}
