package hint

import (
	"net/http"
)

// handlerFunc for users define methods and actions of request path
type handlerFunc func(c *Context)

// Engine implements interface named ServeHTTP
type Engine struct {
	router *router
}

// New is the constructor of Engine for users
func New() *Engine {
	return &Engine{router: newRouter()}
}

// inside func for users to add router
// m -> http method(get/post)
// p -> path
// h -> handler func
func (e *Engine) addRouter(m string, p string, h handlerFunc) {
	e.router.addRouter(m, p, h)
}

// Run is a method for users to run the server on appoint port
func (e *Engine) Run(port string) error {
	return http.ListenAndServe(port, e)
}

// GET is a method for users to add "get" router
func (e *Engine) GET(p string, h handlerFunc) {
	e.addRouter("GET", p, h)
}

// POST is a method for users to add "post" router
func (e *Engine) POST(p string, h handlerFunc) {
	e.addRouter("POST", p, h)
}

// impl interface named ServeHTTP
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	e.router.handle(c)
}
