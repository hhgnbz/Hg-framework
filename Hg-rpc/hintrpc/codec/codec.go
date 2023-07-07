package codec

import "io"

type Header struct {
	ServiceMethod string // 调用方法 e.g."User.GetUser"
	Seq           uint64 // 请求序号，可以理解为id，用于区分不同请求
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(h *Header) error
	ReadBody(body any) error
	Write(h *Header, body any) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
	// TODO Json格式编码
	// NewCodecFuncMap[JsonType] = NewJsonCodec
}
