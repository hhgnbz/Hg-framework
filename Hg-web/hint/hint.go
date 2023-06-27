package hint

import (
	"net/http"
)

// handlerFunc for users define methods and actions of request path
type handlerFunc func(c *Context)

// 路由分组
// e.g.
// 以/post开头的路由匿名可访问。
// 以/admin开头的路由需要鉴权。
// 以/api开头的路由是 RESTful 接口，可以对接第三方平台，需要三方平台鉴权。
// 实现的分组控制也是以前缀来区分，并且支持分组的嵌套。
// e.g.
// /post是一个分组，/post/a和/post/b可以是该分组下的子分组。
// 作用在/post分组上的中间件(middleware)，也都会作用在子分组，子分组还可以应用自己特有的中间件。

type RouterGroup struct {
	prefix      string
	parent      *RouterGroup // 支持分组嵌套
	middlewares []handlerFunc
	engine      *Engine // 所有分组共享一个Engine，保存一个指针方便通过Engine访问其他接口
}

// Engine implements interface named ServeHTTP
type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup // 保存所有分组，所有路由操作通过分组实现
}

// New is the constructor of Engine for users
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Group is defined to create a new RouterGroup
// remember all groups share the same Engine instance
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	e := group.engine
	newGroup := &RouterGroup{
		engine: e,
		prefix: group.prefix + prefix,
		parent: group,
	}
	e.groups = append(e.groups, newGroup)
	return newGroup
}

// inside func for users to add router
// m -> http method(get/post)
// p -> path
// h -> handler func
func (group *RouterGroup) addRouter(m string, p string, h handlerFunc) {
	pattern := group.prefix + p
	group.engine.router.addRouter(m, pattern, h)
}

// Run is a method for users to run the server on appoint port
func (e *Engine) Run(port string) error {
	return http.ListenAndServe(port, e)
}

// GET is a method for users to add "get" router
func (group *RouterGroup) GET(p string, h handlerFunc) {
	group.addRouter("GET", p, h)
}

// POST is a method for users to add "post" router
func (group *RouterGroup) POST(p string, h handlerFunc) {
	group.addRouter("POST", p, h)
}

// impl interface named ServeHTTP
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	e.router.handle(c)
}
