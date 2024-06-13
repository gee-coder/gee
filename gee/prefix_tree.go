package gee

import "strings"

const SEPARATOR = "/"

type treeNode struct {
	nodeName   string
	children   []*treeNode
	routerName string
	isEnd      bool
}

// Put path: /user/get/:id
func (t *treeNode) Put(path string) {
	root := t
	strs := strings.Split(path, SEPARATOR)
	for index, nodeName := range strs {
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, node := range children {
			if node.nodeName == nodeName {
				isMatch = true
				t = node
				break
			}
		}
		if !isMatch {
			isEnd := false
			if index == len(strs)-1 {
				isEnd = true
			}
			node := &treeNode{nodeName: nodeName, children: make([]*treeNode, 0), routerName: t.routerName + SEPARATOR + nodeName, isEnd: isEnd}
			children = append(children, node)
			t.children = children
			t = node
		}
	}
	t = root
}

// Get path: /user/get/1
// /hello
func (t *treeNode) Get(path string) *treeNode {
	nodeNames := strings.Split(path, SEPARATOR)
	for index, nodeName := range nodeNames {
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, node := range children {
			if node.nodeName == nodeName ||
				node.nodeName == "*" ||
				strings.Contains(node.nodeName, ":") {
				isMatch = true
				if index == len(nodeNames)-1 {
					return node
				} else {
					t = node
				}
				break
			}
		}
		if !isMatch {
			for _, node := range children {
				// /user/**
				// /user/get/userInfo
				// /user/aa/bb
				if node.nodeName == "**" {
					return node
				}
			}

		}
	}
	return nil
}
