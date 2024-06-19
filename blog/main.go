package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gee-coder/gee"
	geeError "github.com/gee-coder/gee/error"
	geeLog "github.com/gee-coder/gee/log"
)

type User struct {
	Name      string   `xml:"name" json:"name"`
	Age       int      `xml:"age" json:"age" validate:"required,max=50,min=18"`
	Sex       int      `xml:"sex" json:"sex"`
	Phone     string   `xml:"phone" json:"phone"`
	Email     string   `xml:"email" json:"email" gee:"required"`
	Addresses []string `xml:"addresses" json:"addresses"`
}

func a(param int, geeError *geeError.GeeError) {
	if param == 1 {
		// 发生错误的时候，放入一个地方 然后进行统一处理
		err := errors.New("a error")
		geeError.Put(err)
	}
}

func b(param int, geeError *geeError.GeeError) {
	if param == 1 {
		err2 := errors.New("b error")
		geeError.Put(err2)
	}
}

func c(param int, geeError *geeError.GeeError) {
	if param == 1 {
		err2 := errors.New("c error")
		geeError.Put(err2)
	}
}

type BlogResponse struct {
	Success bool
	Code    int
	Data    any
	Msg     string
}

type BlogNoDataResponse struct {
	Success bool
	Code    int
	Msg     string
}

func (b *BlogResponse) Error() string {
	return b.Msg
}

func (b *BlogResponse) Response() any {
	if b.Data == nil {
		return &BlogNoDataResponse{
			Success: false,
			Code:    -999,
			Msg:     "账号密码错误",
		}
	}
	return b
}

func login() *BlogResponse {
	return &BlogResponse{
		Success: false,
		Code:    -999,
		Data:    nil,
		Msg:     "账号密码错误",
	}
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
	engine := gee.Default()

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

	// query参数
	group.Get("/add", func(ctx *gee.Context) {
		fmt.Println("handler")
		id := ctx.GetQuery("id")
		fmt.Println(id)
	})
	group.Get("/add2", func(ctx *gee.Context) {
		fmt.Println("handler")
		id := ctx.DefaultQuery("id", "-1")
		fmt.Println(id)
	})

	// map
	group.Get("/queryMap", func(ctx *gee.Context) {
		m, _ := ctx.GetQueryMap("user")
		ctx.JSON(http.StatusOK, m)
	})

	// form
	group.Post("/formPost", func(ctx *gee.Context) {
		m, _ := ctx.GetPostFormMap("user")
		files := ctx.FormFiles("file")
		for _, file := range files {
			ctx.SaveUploadedFile(file, "./upload/"+file.Filename)
		}
		ctx.JSON(http.StatusOK, m)
	})

	// json
	group.Post("/jsonParam", func(ctx *gee.Context) {
		user := &User{}
		ctx.DisallowUnknownFields = true
		ctx.IsValidate = true
		err := ctx.BindJson(user)
		if err == nil {
			ctx.JSON(http.StatusOK, user)
		} else {
			log.Println(err)
		}
	})
	group.Post("/jsonParam2", func(ctx *gee.Context) {
		user := make([]User, 0)
		ctx.DisallowUnknownFields = true
		ctx.IsValidate = true
		err := ctx.BindJson(&user)
		if err == nil {
			ctx.JSON(http.StatusOK, user)
		} else {
			log.Println(err)
		}
	})

	engine.Logger.Level = geeLog.LevelInfo
	// engine.Logger.Formatter = &geeLog.JsonFormatter{TimeDisplay: true}
	// engine.Logger.SetLogPath("./log")
	// defer engine.Logger.CloseWriter()

	group.Post("/xmlParam1", func(ctx *gee.Context) {
		user := &User{}
		_ = ctx.BindXML(user)
		err := errors.New("a error")
		if err != nil {
			panic(err)
		}
		if err == nil {
			ctx.JSON(http.StatusOK, user)
		} else {
			log.Println(err)
		}
	})

	group.Post("/xmlParam2", func(ctx *gee.Context) {
		user := &User{}
		_ = ctx.BindXML(user)
		err := geeError.Default()
		err.Result(func(geeError *geeError.GeeError) {
			ctx.Logger.Info(geeError.Error())
			ctx.JSON(http.StatusInternalServerError, user)
		})
		a(1, err)
		b(1, err)
		c(1, err)
		ctx.JSON(http.StatusOK, user)
	})

	group.Post("/xmlParam3", func(ctx *gee.Context) {
		user := &User{}
		_ = ctx.BindXML(user)
		err := geeError.Default()
		err.Result(func(geeError *geeError.GeeError) {
			ctx.Logger.Info(geeError.Error())
			ctx.JSON(http.StatusInternalServerError, user)
		})
		a(1, err)
		b(1, err)
		c(1, err)
		ctx.JSON(http.StatusOK, user)
	})

	engine.RegisterErrorHandler(func(err error) (int, any) {
		switch e := err.(type) {
		case *BlogResponse:
			return http.StatusOK, e.Response()
		default:
			return http.StatusInternalServerError, "Internal Server Error"
		}
	})

	group.Post("/xmlParam4", func(ctx *gee.Context) {
		user := &User{}
		_ = ctx.BindXML(user)
		err := login()
		ctx.HandleWithError(http.StatusOK, user, err)
	})

	engine.Run()
}
