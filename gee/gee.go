package gee

import (
	"log"
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

type router struct {
	groups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	g := &routerGroup{groupName: name, handlerMap: make(map[string]HandlerFunc)}
	r.groups = append(r.groups, g)
	return g
}

func (r *routerGroup) Add(name string, handlerFunc HandlerFunc) {
	r.handlerMap[name] = handlerFunc
}

type routerGroup struct {
	groupName  string
	handlerMap map[string]HandlerFunc
}

type Engine struct {
	*router
}

func New() *Engine {
	return &Engine{
		&router{},
	}
}

func (e *Engine) Run() {
	groups := e.router.groups
	for _, g := range groups {
		for name, handle := range g.handlerMap {
			http.HandleFunc("/"+g.groupName+name, handle)
		}
	}
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
