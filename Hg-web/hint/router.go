package hint

import (
	"log"
	"net/http"
)

// 将路由相关的方法和结构提取出来，方便对 router 的功能进行增强
// 例如，提供动态路由的支持

// router struct
type router struct {
	handlers map[string]handlerFunc
}

// constructor of Router
func newRouter() *router {
	return &router{
		handlers: make(map[string]handlerFunc),
	}
}

// handle router
func (r *router) handle(c *Context) {
	k := c.Method + "-" + c.Path
	if handler, ok := r.handlers[k]; ok {
		handler(c)
	} else {
		c.String(http.StatusNotFound, "404 Not Found: %s\n", c.Path)
	}
}

// inside func for users to add router
// m -> http method(get/post)
// p -> path
// h -> handler func
func (r *router) addRouter(m string, p string, h handlerFunc) {
	log.Printf("Route %4s - %s", m, p)
	k := m + "-" + p
	r.handlers[k] = h
}
