package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gee-coder/gee"
)

type User struct {
	Name string `xml:"name"`
	Age  int    `xml:"age"`
}

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
		_, err := fmt.Fprintln(ctx.W, "user/hello Get geecoder.net")
		if err != nil {
			log.Println(err)
		}
	})
	group.Get("/hello/*/get", func(ctx *gee.Context) {
		fmt.Println("handler")
		_, err := fmt.Fprintln(ctx.W, "/hello/*/get Get geecoder.net")
		if err != nil {
			log.Println(err)
		}
	})
	group.Post("/hello", func(ctx *gee.Context) {
		fmt.Println("handler")
		_, err := fmt.Fprintln(ctx.W, "user/hello Post geecoder.net")
		if err != nil {
			log.Println(err)
		}
	})
	group.Post("/info", func(ctx *gee.Context) {
		fmt.Println("handler")
		_, err := fmt.Fprintln(ctx.W, "user/info Post geecoder.net")
		if err != nil {
			log.Println(err)
		}
	})
	group.Any("/any", func(ctx *gee.Context) {
		fmt.Println("handler")
		_, err := fmt.Fprintln(ctx.W, "user/any Any geecoder.net")
		if err != nil {
			log.Println(err)
		}
	})
	// user/get/2
	group.Get("/get/:id", func(ctx *gee.Context) {
		fmt.Println("handler")
		_, err := fmt.Fprintln(ctx.W, "/get/:id Get geecoder.net")
		if err != nil {
			log.Println(err)
		}
	}, Log)

	// string
	group.Get("/string", func(ctx *gee.Context) {
		fmt.Println("handler")
		_ = ctx.String(http.StatusOK, "%s 是由 %s 制作 \n", "gee框架", "geecoder")
	})

	// xml
	group.Get("/xml", func(ctx *gee.Context) {
		fmt.Println("handler")
		user := &User{
			Name: "牛牛",
			Age:  18,
		}
		_ = ctx.XML(http.StatusOK, user)
	})

	// json
	group.Get("/json", func(ctx *gee.Context) {
		fmt.Println("handler")
		_ = ctx.JSON(http.StatusOK, &User{
			Name: "牛牛",
			Age:  18,
		})
	})

	// html
	group.Get("/html", func(ctx *gee.Context) {
		fmt.Println("handler")
		_ = ctx.HTML(http.StatusOK, "<h1>你好 Html</h1>")
	})
	group.Get("/index", func(ctx *gee.Context) {
		fmt.Println("handler")
		ctx.HTMLTemplate("index.html", template.FuncMap{}, "", "tpl/index.html")
	})
	group.Get("/htmlTemplate", func(ctx *gee.Context) {
		fmt.Println("handler")
		user := &User{
			Name: "牛牛",
		}
		ctx.HTMLTemplate("login.html", template.FuncMap{}, user, "tpl/login.html", "tpl/header.html")
	})
	group.Get("/login", func(ctx *gee.Context) {
		fmt.Println("handler")
		user := &User{
			Name: "牛牛",
		}
		ctx.HTMLTemplateGlob("login.html", template.FuncMap{}, user, "tpl/*.html")
	})
	// 预加载
	engine.LoadTemplate("tpl/*.html")
	engine.SetFuncMap(template.FuncMap{})
	group.Get("/login2", func(ctx *gee.Context) {
		fmt.Println("handler")
		user := &User{
			Name: "牛牛2",
		}
		err := ctx.TemplateHTML("login.html", user)
		if err != nil {
			log.Println(err)
		}
	})

	// 重定向页面
	group.Get("/redirect", func(ctx *gee.Context) {
		fmt.Println("handler")
		_ = ctx.Redirect(http.StatusFound, "/user/login")
	})

	// file
	group.Get("/file", func(ctx *gee.Context) {
		fmt.Println("handler")
		ctx.File("tpl/test.xlsx")
	})
	// 指定文件名
	group.Get("/file/name", func(ctx *gee.Context) {
		fmt.Println("handler")
		ctx.FileAttachment("tpl/test.xlsx", "牛牛.xlsx")
	})
	group.Get("/file/fs", func(ctx *gee.Context) {
		fmt.Println("handler")
		ctx.FileFromFS("test.xlsx", http.Dir("tpl"))
	})

	engine.Run()
}
