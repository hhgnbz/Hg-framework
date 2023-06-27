package main

import (
	"fmt"
	"hint"
	"html/template"
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

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := hint.Default()
	r.Use(hint.Logger())
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	stu1 := &student{Name: "hg", Age: 24}
	stu2 := &student{Name: "axg", Age: 23}
	r.GET("/", func(c *hint.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/students", func(c *hint.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", hint.H{
			"title":  "hg",
			"stuArr": [2]*student{stu1, stu2},
		})
	})

	r.GET("/date", func(c *hint.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", hint.H{
			"title": "hg",
			"now":   time.Date(2023, 6, 27, 0, 0, 0, 0, time.UTC),
		})
	})

	// index out of range for testing Recovery()
	r.GET("/panic", func(c *hint.Context) {
		names := []string{"hg"}
		c.String(http.StatusOK, names[100])
	})

	r.Run(":9999")
}
