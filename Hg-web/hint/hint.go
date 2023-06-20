package hint

import (
	"fmt"
	"net/http"
)

// handlerFunc for users define methods and actions of request path
type handlerFunc func(w http.ResponseWriter, req *http.Request)

// Engine implements interface named ServeHTTP
type Engine struct {
	router map[string]handlerFunc
}

// New is the constructor of Engine for users
func New() *Engine {
	return &Engine{make(map[string]handlerFunc)}
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
	k := req.Method + "-" + req.URL.Path
	if handler, ok := e.router[k]; ok {
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 Not Found:%q\n", req.URL.Path)
	}
}

// inside func for users to add router
// m -> http method(get/post)
// p -> path
// h -> handler func
func (e *Engine) addRouter(m string, p string, h handlerFunc) {
	k := m + "-" + p
	e.router[k] = h
}
