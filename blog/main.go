package main

import (
	"fmt"

	"github.com/gee-coder/gee"
)

func Log(next gee.HandlerFunc) gee.HandlerFunc {
	return func(ctx *gee.Context) {
		fmt.Println("func Log 打印请求参数")
		// 执行本身的代码
		next(ctx)
		fmt.Println("func Log 返回执行时间")
	}
}

func main() {
	engine := gee.New()

	group := engine.Group("user")
	group.AddMiddlewareFunc(func(next gee.HandlerFunc) gee.HandlerFunc {
		return func(ctx *gee.Context) {
			fmt.Println("执行前置中间件代码")
			// 执行本身的代码
			next(ctx)
			fmt.Println("执行后置中间件代码")
		}
	})
	group.Get("/hello", func(ctx *gee.Context) {
		fmt.Println("handler")
		fmt.Fprintln(ctx.W, "user/hello Get geecoder.net")
	})
	group.Get("/hello/*/get", func(ctx *gee.Context) {
		fmt.Println("handler")
		fmt.Fprintln(ctx.W, "/hello/*/get Get geecoder.net")
	})
	group.Post("/hello", func(ctx *gee.Context) {
		fmt.Println("handler")
		fmt.Fprintln(ctx.W, "user/hello Post geecoder.net")
	})
	group.Post("/info", func(ctx *gee.Context) {
		fmt.Println("handler")
		fmt.Fprintln(ctx.W, "user/info Post geecoder.net")
	})
	group.Any("/any", func(ctx *gee.Context) {
		fmt.Println("handler")
		fmt.Fprintln(ctx.W, "user/any Any geecoder.net")
	})
	// user/get/2
	group.Get("/get/:id", func(ctx *gee.Context) {
		fmt.Println("handler")
		fmt.Fprintln(ctx.W, "/get/:id Get geecoder.net")
	}, Log)

	engine.Run()
}
