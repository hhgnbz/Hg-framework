package main

import (
	"hint"
	"log"
	"net/http"
	"time"
)

func onlyForV2() hint.HandlerFunc {
	return func(c *hint.Context) {
		// Start timer
		t := time.Now()
		// if a server error occurred
		c.Fail(500, "Internal Server Error")
		// Calculate resolution time
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func main() {
	r := hint.New()
	r.Use(hint.Logger()) // global midlleware
	r.GET("/", func(c *hint.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Hint</h1>")
	})

	v2 := r.Group("/v2")
	v2.Use(onlyForV2()) // v2 group middleware
	{
		v2.GET("/hello/:name", func(c *hint.Context) {
			// expect /hello/hg
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
	}

	r.Run(":9999")
}
