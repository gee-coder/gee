package gee

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/gee-coder/gee/render"
)

// 请求上下文
type Context struct {
	W      http.ResponseWriter
	R      *http.Request
	engine *Engine
}

func (c *Context) HTMLTemplate(name string, funcMap template.FuncMap, data any, fileName ...string) {
	t := template.New(name)
	t.Funcs(funcMap)
	t, err := t.ParseFiles(fileName...)
	if err != nil {
		log.Println(err)
		return
	}
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Execute(c.W, data)
	if err != nil {
		log.Println(err)
	}
}

func (c *Context) HTMLTemplateGlob(name string, funcMap template.FuncMap, data any, pattern string) {
	t := template.New(name)
	t.Funcs(funcMap)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		log.Println(err)
		return
	}
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Execute(c.W, data)
	if err != nil {
		log.Println(err)
	}
}

func (c *Context) Render(r render.Render, code int) error {
	return r.Render(c.W, code)
}

func (c *Context) String(code int, format string, values ...any) error {
	return c.Render(&render.String{
		Format: format,
		Data:   values,
	}, code)
}

func (c *Context) XML(code int, data any) error {
	return c.Render(&render.XML{Data: data}, code)
}

func (c *Context) JSON(code int, data any) error {
	return c.Render(&render.JSON{Data: data}, code)
}

func (c *Context) HTML(code int, html string) error {
	return c.Render(&render.HTML{IsTemplate: false, Data: html}, code)
}

func (c *Context) TemplateHTML(name string, data any) error {
	return c.Render(&render.HTML{
		Name:       name,
		IsTemplate: true,
		Template:   c.engine.HTMLRender.Template,
		Data:       data,
	}, http.StatusOK)
}

// 重定向页面
func (c *Context) Redirect(code int, location string) error {
	return c.Render(&render.Redirect{
		Code:     code,
		Request:  c.R,
		Location: location,
	}, http.StatusOK)
}

// 下载文件
func (c *Context) File(filePath string) {
	http.ServeFile(c.W, c.R, filePath)
}

// 指定文件名
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

// 指定文件系统路径 filepath是指定文件系统路径下的路径
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	// 恢复路径
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)
	c.R.URL.Path = filepath
	http.FileServer(fs).ServeHTTP(c.W, c.R)
}
