package main

import (
	"hint"
	"net/http"
)

func main() {
	r := hint.New()
	r.GET("/index", func(c *hint.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	v1 := r.Group("/v1")
	{
		v1.GET("/", func(c *hint.Context) {
			c.HTML(http.StatusOK, "<h1>Hello Hint</h1>")
		})

		v1.GET("/hello", func(c *hint.Context) {
			// expect /hello?name=hg
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *hint.Context) {
			// expect /hello/hg
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *hint.Context) {
			c.JSON(http.StatusOK, hint.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})

	}

	r.Run(":9999")
}
