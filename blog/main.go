package main

import (
	"fmt"
	"github.com/gee-coder/gee"
	"net/http"
)

func main() {
	engine := gee.New()
	group := engine.Group("user")
	group.Add("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello geecoder.net")
	})

	engine.Run()
}
