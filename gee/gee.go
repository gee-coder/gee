package gee

import (
	"fmt"
	"log"
	"net/http"
)

const ANY = "ANY"

type Context struct {
	W http.ResponseWriter
	R *http.Request
}

type HandlerFunc func(ctx *Context)

type routerGroup struct {
	// 组
	groupName string
	// key1:组下的路由 key2:请求类型 value:处理方法 用来保存处理方法
	handlerMap map[string]map[string]HandlerFunc
	// key:组下的路由 value:请求类型数组 用来判断[ip:端口：/组+组下的路由]是否支持此种请求类型
	handlerMethodMap map[string][]string
	// 前缀树 每个组下都有一个前缀树 用来匹配具体的请求路径
	treeNode *treeNode
}

func (r *routerGroup) handle(routerName string, method string, handlerFunc HandlerFunc) {
	_, ok := r.handlerMap[routerName]
	if !ok {
		r.handlerMap[routerName] = make(map[string]HandlerFunc)
	}
	r.handlerMap[routerName][method] = handlerFunc
	r.handlerMethodMap[method] = append(r.handlerMethodMap[method], routerName)
	methodMap := make(map[string]HandlerFunc)
	methodMap[method] = handlerFunc
	r.treeNode.Put(routerName)
}

func (r *routerGroup) Handle(routerName string, method string, handlerFunc HandlerFunc) {
	// method有效性做校验
	r.handle(routerName, method, handlerFunc)
}

func (r *routerGroup) Any(routerName string, handlerFunc HandlerFunc) {
	r.handle(routerName, ANY, handlerFunc)
}

func (r *routerGroup) Get(routerName string, handlerFunc HandlerFunc) {
	r.handle(routerName, http.MethodGet, handlerFunc)
}

func (r *routerGroup) Post(routerName string, handlerFunc HandlerFunc) {
	r.handle(routerName, http.MethodPost, handlerFunc)
}

func (r *routerGroup) Delete(routerName string, handlerFunc HandlerFunc) {
	r.handle(routerName, http.MethodDelete, handlerFunc)
}

func (r *routerGroup) Put(routerName string, handlerFunc HandlerFunc) {
	r.handle(routerName, http.MethodPut, handlerFunc)
}

func (r *routerGroup) Patch(routerName string, handlerFunc HandlerFunc) {
	r.handle(routerName, http.MethodPatch, handlerFunc)
}

func (r *routerGroup) Options(routerName string, handlerFunc HandlerFunc) {
	r.handle(routerName, http.MethodOptions, handlerFunc)
}

func (r *routerGroup) Head(routerName string, handlerFunc HandlerFunc) {
	r.handle(routerName, http.MethodHead, handlerFunc)
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	g := &routerGroup{
		groupName:        name,
		handlerMap:       make(map[string]map[string]HandlerFunc),
		handlerMethodMap: make(map[string][]string),
		treeNode:         &treeNode{nodeName: SEPARATOR, children: make([]*treeNode, 0)},
	}
	r.routerGroups = append(r.routerGroups, g)
	return g
}

type Engine struct {
	router
}

func New() *Engine {
	return &Engine{
		router{},
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	for _, group := range e.routerGroups {
		routerName := SubStringLast(r.RequestURI, "/"+group.groupName)
		// get/1
		node := group.treeNode.Get(routerName)
		if node != nil && node.isEnd {
			// 路由匹配上了
			ctx := &Context{
				W: w,
				R: r,
			}
			handle, ok := group.handlerMap[node.routerName][ANY]
			if ok {
				handle(ctx)
				return
			}
			handle, ok = group.handlerMap[node.routerName][method]
			if ok {
				handle(ctx)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "%s %s not allowed \n", r.RequestURI, method)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s  not found \n", r.RequestURI)
}

func (e *Engine) Run() {
	http.Handle(SEPARATOR, e)
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
