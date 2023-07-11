package hintrpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

// 对 net/rpc 而言，一个函数需要能够被远程调用，需要满足五个条件
// 1.方法类型对外暴露
// 2.方法对外暴露
// 3.方法有2个参数，均对外暴露
// 4.方法第2个参数为指针
// 5.方法有1个error返回值
// e.g.
// func (t *T) MethodName(argType T1, replyType *T2) error
// 假设客户端发来一个请求，包含ServiceMethod和Argv
// {
//    "ServiceMethod"： "T.MethodName"
//    "Argv"："0101110101..." // 序列化之后的字节流
// }
// 硬编码实现: switch case...
// 映射过程自动化 -> 反射reflect
// reflect.TypeOf() 获取结构体所有方法

type methodType struct {
	method    reflect.Method // 方法本身
	ArgType   reflect.Type   // 第一个参数类型，可以是值类型，也可以是指针
	ReplyType reflect.Type   // 第二个参数类型，只能是指针
	numCalls  uint64         // 统计方法调用次数
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() (argv reflect.Value) {
	// arg may be a pointer type,or a value type
	if m.ArgType.Kind() == reflect.Ptr {
		// 指针类型
		argv = reflect.New(m.ArgType.Elem())
	} else {
		// 值类型
		argv = reflect.New(m.ArgType).Elem()
	}
	return
}

func (m *methodType) newReplyv() (replyv reflect.Value) {
	// reply must be a pointer type
	replyv = reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return
}

type service struct {
	name   string                 // 映射结构体名称
	typ    reflect.Type           // 结构体类型
	rcvr   reflect.Value          // 结构体实例本身，调用方法时需要rcvr作为第0个参数
	method map[string]*methodType // 存储映射结构体所有符合条件方法
}

func newService(rcvr any) (s *service) {
	s = new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return
}

// registerMethods 注册service中所有符合rpc/net远程调用的方法
func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		// 检验参数合法
		// 入参为3个，rcvr/args/reply,出参为1个，error
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		// 出参不为error时
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		// args,reply必须为暴露或内建类型
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

// isExportedOrBuiltinType 判断类型是否为暴露或内建类型
func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// call 通过反射值调用方法
func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	// 增加调用次数，后续统计调用次数使用
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
