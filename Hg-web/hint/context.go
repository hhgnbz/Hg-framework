package hint

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 对Web服务来说，无非是根据请求*http.Request，构造响应http.ResponseWriter。但是这两个对象提供的接口粒度太细。
// 比如我们要构造一个完整的响应，需要考虑消息头(Header)和消息体(Body)。
// 而 Header 包含了状态码(StatusCode)，消息类型(ContentType)等几乎每次请求都需要设置的信息。
// 因此，如果不进行有效的封装，那么框架的用户将需要写大量重复，繁杂的代码，而且容易出错。
// 针对常用场景，能够高效地构造出 HTTP 响应是一个好的框架必须考虑的点。
// 对于框架来说，context提供额外的支撑功能。
// 例如，当前请求中解析动态路由参数的存放、中间件产生的信息。

// H for users to make JSON text
type H map[string]interface{}

// Context struct
type Context struct {
	// origin info
	Writer http.ResponseWriter
	Req    *http.Request
	// high freq use request info
	Path   string
	Method string
	// high freq use response info
	StatusCode int
}

// inside func to make a newContext
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

// =========== request part start ===========

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// =========== request part end ===========

// =========== response part start ===========

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

// =========== response part end ===========
