package gee

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"

	geeLog "github.com/gee-coder/gee/log"
	"github.com/gee-coder/gee/render"
)

const ANY = "ANY"

// 方法
type HandlerFunc func(ctx *Context)

// 中间件
type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc

// 路由组
type routerGroup struct {
	// 组
	groupName string
	// key1:组下的路由 key2:请求类型 value:处理方法 用来保存处理方法
	handlerMap map[string]map[string]HandlerFunc
	// key:组下的路由 value:请求类型数组 用来判断[ip:端口：/组+组下的路由]是否支持此种请求类型
	handlerMethodMap map[string][]string
	// 前缀树 每个组下都有一个前缀树 用来匹配具体请求路径
	treeNode *treeNode
	// 组中间件
	middlewares []MiddlewareFunc
	// 路由中间件
	middlewaresFuncMap map[string]map[string][]MiddlewareFunc
}

func (r *routerGroup) handle(routerName string, method string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	_, ok := r.handlerMap[routerName]
	if !ok {
		r.handlerMap[routerName] = make(map[string]HandlerFunc)
		r.middlewaresFuncMap[routerName] = make(map[string][]MiddlewareFunc)
	}
	if r.handlerMap[routerName][method] == nil {
		r.handlerMap[routerName][method] = handlerFunc
	} else {
		panic("[路由：" + routerName + "，方法：" + method + "]已经被注册")
	}
	r.handlerMethodMap[method] = append(r.handlerMethodMap[method], routerName)
	// 组装前缀树
	r.treeNode.Put(routerName)
	// 组装中间件

	// 组装路由中间件
	r.middlewaresFuncMap[routerName][method] = append(r.middlewaresFuncMap[routerName][method], middlewareFunc...)
}

func (r *routerGroup) AddMiddlewareFunc(middlewares ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *routerGroup) methodHandle(routerName string, method string, h HandlerFunc, ctx *Context) {
	// 通用的组中间件
	if r.middlewares != nil {
		// 包裹n层中间件
		for _, middlewareFunc := range r.middlewares {
			h = middlewareFunc(h)
		}
	}
	// 路由级别的组中间件
	if r.middlewaresFuncMap[routerName][method] != nil {
		// 包裹n层中间件
		for _, middlewareFunc := range r.middlewaresFuncMap[routerName][method] {
			h = middlewareFunc(h)
		}
	}
	h(ctx)
}

func (r *routerGroup) Any(routerName string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(routerName, ANY, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Get(routerName string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(routerName, http.MethodGet, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Post(routerName string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(routerName, http.MethodPost, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Delete(routerName string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(routerName, http.MethodDelete, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Put(routerName string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(routerName, http.MethodPut, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Patch(routerName string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(routerName, http.MethodPatch, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Options(routerName string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(routerName, http.MethodOptions, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) Head(routerName string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(routerName, http.MethodHead, handlerFunc, middlewareFunc...)
}

// 路由
type router struct {
	routerGroups []*routerGroup
	engine       *Engine
}

func (r *router) Group(name string) *routerGroup {
	g := &routerGroup{
		groupName:          name,
		handlerMap:         make(map[string]map[string]HandlerFunc),
		handlerMethodMap:   make(map[string][]string),
		treeNode:           &treeNode{nodeName: SEPARATOR, children: make([]*treeNode, 0)},
		middlewares:        make([]MiddlewareFunc, 0),
		middlewaresFuncMap: make(map[string]map[string][]MiddlewareFunc),
	}
	g.AddMiddlewareFunc(r.engine.middlewares...)
	r.routerGroups = append(r.routerGroups, g)
	return g
}

type ErrorHandler func(err error) (int, any)

// 引擎
type Engine struct {
	router
	funcMap    template.FuncMap
	HTMLRender render.HTMLRender
	// sync.Pool用于存储分配了还没被使用但未来可能被使用的值
	// sync.Pool大小可伸缩，会动态扩容，池中不活跃的对象会被自动清理
	pool   sync.Pool
	Logger *geeLog.Logger
	// 全局中间件
	middlewares  []MiddlewareFunc
	errorHandler ErrorHandler
}

func (e *Engine) RegisterErrorHandler(err ErrorHandler) {
	e.errorHandler = err
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) SetHtmlTemplate(t *template.Template) {
	e.HTMLRender = render.HTMLRender{Template: t}
}

// LoadTemplateGlob 加载所有模板
func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.SetHtmlTemplate(t)
}

func (e *Engine) httpRequestHandle(ctx *Context) {
	for _, group := range e.routerGroups {
		routerName := SubStringLast(ctx.R.URL.Path, "/"+group.groupName)
		// get/1
		node := group.treeNode.Get(routerName)
		if node != nil && node.isEnd {
			// 路由匹配上了
			handle, ok := group.handlerMap[node.routerName][ANY]
			if ok {
				group.methodHandle(node.routerName, ANY, handle, ctx)
				return
			}
			handle, ok = group.handlerMap[node.routerName][ctx.R.Method]
			if ok {
				group.methodHandle(node.routerName, ctx.R.Method, handle, ctx)
				return
			}
			ctx.W.WriteHeader(http.StatusMethodNotAllowed)
			_, err := fmt.Fprintf(ctx.W, "%s %s not allowed \n", ctx.R.RequestURI, ctx.R.Method)
			if err != nil {
				log.Println(err)
			}
			return
		}
	}
	ctx.W.WriteHeader(http.StatusNotFound)
	_, err := fmt.Fprintf(ctx.W, "%s  not found \n", ctx.R.RequestURI)
	if err != nil {
		log.Println(err)
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := e.pool.Get().(*Context)
	ctx.W = w
	ctx.R = r
	ctx.Logger = e.Logger
	e.httpRequestHandle(ctx)
	// 存起来可以不用再次分配内存，提高效率
	e.pool.Put(ctx)
}

func (e *Engine) Run() {
	http.Handle(SEPARATOR, e)
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *Engine) allocateContext() any {
	return &Context{engine: e}
}

func (e *Engine) AddMiddlewareFunc(middlewares ...MiddlewareFunc) {
	e.middlewares = append(e.middlewares, middlewares...)
}

func New() *Engine {
	engine := &Engine{
		router:     router{},
		funcMap:    nil,
		HTMLRender: render.HTMLRender{},
		Logger:     geeLog.Default(),
	}
	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	// r.engine = engine
	return engine
}

func Default() *Engine {
	engine := New()
	engine.AddMiddlewareFunc(Recovery, Logging)
	engine.router.engine = engine
	return engine
}
