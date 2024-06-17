package gee

import (
	"errors"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gee-coder/gee/binding"
	"github.com/gee-coder/gee/render"
)

// 32M
const defaultMultipartMemory = 32 << 20

// 请求上下文
type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	engine                *Engine
	queryCache            url.Values
	formCache             url.Values
	DisallowUnknownFields bool
	IsValidate            bool
	StatusCode            int
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
	err := r.Render(c.W, code)
	if code != http.StatusOK {
		c.W.WriteHeader(code)
	}
	return err
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

func (c *Context) initQueryCache() {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

func (c *Context) QueryArray(key string) (values []string) {
	c.initQueryCache()
	values, _ = c.queryCache[key]
	return
}

func (c *Context) GetQueryArray(key string) (values []string, ok bool) {
	c.initQueryCache()
	values, ok = c.queryCache[key]
	return
}

func (c *Context) DefaultQuery(key, defaultValue string) string {
	array, ok := c.GetQueryArray(key)
	if !ok {
		return defaultValue
	}
	return array[0]
}

func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	// user[id]=1&user[name]=张三
	dict := make(map[string]string)
	exist := false
	for k, value := range m {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dict[k[i+1:][:j]] = value[0]
			}
		}
	}
	return dict, exist
}

func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

func (c *Context) QueryMap(key string) (dict map[string]string) {
	dict, _ = c.GetQueryMap(key)
	return
}

func (c *Context) initFormCache() {
	if c.formCache == nil {
		c.formCache = make(url.Values)
		if err := c.R.ParseMultipartForm(defaultMultipartMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
			}
		}
		c.formCache = c.R.PostForm
	}
}

func (c *Context) GetPostFormArray(key string) (values []string, ok bool) {
	c.initFormCache()
	values, ok = c.formCache[key]
	return
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) PostFormArray(key string) (values []string) {
	values, _ = c.GetPostFormArray(key)
	return
}

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initFormCache()
	return c.get(c.formCache, key)
}

func (c *Context) PostFormMap(key string) (dict map[string]string) {
	dict, _ = c.GetPostFormMap(key)
	return
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(defaultMultipartMemory)
	return c.R.MultipartForm, err
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	req := c.R
	if err := req.ParseMultipartForm(defaultMultipartMemory); err != nil {
		return nil, err
	}
	file, header, err := req.FormFile(name)
	if err != nil {
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	return header, nil
}

func (c *Context) FormFiles(name string) []*multipart.FileHeader {
	multipartForm, err := c.MultipartForm()
	if err != nil {
		return make([]*multipart.FileHeader, 0)
	}
	return multipartForm.File[name]
}

func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dstPath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (c *Context) ShouldBindWith(obj any, b binding.Binding) error {
	return b.Bind(c.R, obj)
}

func (c *Context) MustBindWith(obj any, b binding.Binding) error {
	// 如果发生错误，返回400状态码 参数错误
	if err := c.ShouldBindWith(obj, b); err != nil {
		c.W.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}

func (c *Context) BindJson(obj any) error {
	jsonBinding := binding.JSON
	jsonBinding.DisallowUnknownFields = c.DisallowUnknownFields
	jsonBinding.IsValidate = c.IsValidate
	return c.MustBindWith(obj, jsonBinding)
}

func (c *Context) BindXML(obj any) error {
	return c.MustBindWith(obj, binding.XML)
}
