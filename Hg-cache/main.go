package main

import (
	"fmt"
	"hintcache"
	"log"
	"net/http"
)

var db = map[string]string{
	"hhg":  "creator",
	"cow":  "mou",
	"cat":  "meow",
	"ping": "pong",
}

func main() {
	hintcache.NewGroup("test", 2<<10, hintcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	addr := "localhost:9999"
	peers := hintcache.NewHTTPPool(addr)
	log.Println("hintcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
