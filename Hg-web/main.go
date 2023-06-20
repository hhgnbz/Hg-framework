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
		// expect /hello?name=${your name}
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.POST("/login", func(c *hint.Context) {
		c.JSON(http.StatusOK, hint.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})

	r.Run(":9999")
}
