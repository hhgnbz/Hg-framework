package main

import (
	"fmt"
	"hint"
	"log"
	"net/http"
)

func main() {
	//http.HandleFunc("/", indexHandler)
	//http.HandleFunc("/hello", helloHandler)
	e := hint.New()
	e.GET("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})
	log.Fatal(e.Run(":9999"))
}

// handler echos r.URL.Path
func indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Url.Path = %q\n", req.URL.Path)
}

// handler echos r.Header -> k,v
func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Req.Header[%q] = %q\n", k, v)
	}
}
