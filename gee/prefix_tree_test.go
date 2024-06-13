package gee

import (
	"fmt"
	"testing"
)

func TestTreeNode(t *testing.T) {
	root := &treeNode{nodeName: "/", children: make([]*treeNode, 0), routerName: "/", isEnd: false}
	root.Put("/user/get/:id")
	root.Put("/user/create/hello")
	root.Put("/user/create/aaa")
	root.Put("/order/get/aaa")

	node := root.Get("/user/get/:id")
	fmt.Println(node)
	node = root.Get("/user/create/hello")
	fmt.Println(node)
	node = root.Get("/user/create/aaa")
	fmt.Println(node)
	node = root.Get("/order/get/aaa")
	fmt.Println(node)
}
