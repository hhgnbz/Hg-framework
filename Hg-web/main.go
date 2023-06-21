package main

import (
	"hint"
	"net/http"
)

func main() {
	r := hint.New()
	r.GET("/", func(c *hint.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})

	r.GET("/hello", func(c *hint.Context) {
		// expect /hello?name=hb
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.GET("/hello/:name", func(c *hint.Context) {
		// expect /hello/hb
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})

	r.GET("/assets/*filepath", func(c *hint.Context) {
		c.JSON(http.StatusOK, hint.H{"filepath": c.Param("filepath")})
	})

	r.Run(":9999")
}
