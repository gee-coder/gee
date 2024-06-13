package main

import (
	"fmt"

	"github.com/gee-coder/gee"
)

func main() {
	engine := gee.New()

	group := engine.Group("user")
	group.Get("/hello", func(ctx *gee.Context) {
		fmt.Fprintln(ctx.W, "user/hello Get geecoder.net")
	})
	group.Get("/hello/*/get", func(ctx *gee.Context) {
		fmt.Fprintln(ctx.W, "/hello/*/get Get geecoder.net")
	})
	group.Post("/hello", func(ctx *gee.Context) {
		fmt.Fprintln(ctx.W, "user/hello Post geecoder.net")
	})
	group.Post("/info", func(ctx *gee.Context) {
		fmt.Fprintln(ctx.W, "user/info Post geecoder.net")
	})
	group.Any("/any", func(ctx *gee.Context) {
		fmt.Fprintln(ctx.W, "user/any Any geecoder.net")
	})
	// user/get/2
	group.Get("/get/:id", func(ctx *gee.Context) {
		fmt.Fprintln(ctx.W, "/get/:id Get geecoder.net")
	})

	engine.Run()
}
