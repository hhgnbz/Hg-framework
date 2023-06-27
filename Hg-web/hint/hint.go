package hint

import (
	"html/template"
	"net/http"
	"path"
	"strings"
)

// HandlerFunc for users define methods and actions of request path
type HandlerFunc func(c *Context)

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
	middlewares []HandlerFunc
	engine      *Engine // 所有分组共享一个Engine，保存一个指针方便通过Engine访问其他接口
}

// Engine implements interface named ServeHTTP
type Engine struct {
	*RouterGroup
	router        *router
	groups        []*RouterGroup     // 保存所有分组，所有路由操作通过分组实现
	htmlTemplates *template.Template // 模板加载进内存
	funcMap       template.FuncMap   // 所有的自定义模板渲染函数
}

// SetFuncMap method for users to use
func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

// LoadHTMLGlob method
func (e *Engine) LoadHTMLGlob(pattern string) {
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}

// New is the constructor of Engine for users
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Default use Logger() & Recovery middlewares
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
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
func (group *RouterGroup) addRouter(m string, p string, h HandlerFunc) {
	pattern := group.prefix + p
	group.engine.router.addRouter(m, pattern, h)
}

// Run is a method for users to run the server on appoint port
func (e *Engine) Run(port string) error {
	return http.ListenAndServe(port, e)
}

// GET is a method for users to add "get" router
func (group *RouterGroup) GET(p string, h HandlerFunc) {
	group.addRouter("GET", p, h)
}

// POST is a method for users to add "post" router
func (group *RouterGroup) POST(p string, h HandlerFunc) {
	group.addRouter("POST", p, h)
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// impl interface named ServeHTTP
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	middlewares := make([]HandlerFunc, 0)
	for _, g := range e.groups {
		if strings.HasPrefix(req.URL.Path, g.prefix) {
			middlewares = append(middlewares, g.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.e = e
	e.router.handle(c)
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static files method for users to use
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	group.GET(urlPattern, handler)
}
