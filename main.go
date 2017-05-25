package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rovarghe/mule/plugin"
)

type H struct {
}

func (h *H) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RequestURI, "served")

	switch r.URL.Path {
	case "/":
		fmt.Println("RtmpOot served")
	default:
		http.DefaultServeMux.ServeHTTP(w, r)
	}
}

func startServer() error {
	return http.ListenAndServe(":8000", &H{})
}

func main() {
	plugin.TestPlugin()
	log.Fatal(startServer())
}
