package service

import (
	geeRpc "github.com/gee-coder/gee/rpc"
)

type Goods struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type GoodsService struct {
	Find func(args map[string]any) ([]byte, error) `geerpc:"GET,/goods/find"`
}

func (r *GoodsService) Env() geeRpc.HttpConfig {
	c := geeRpc.HttpConfig{
		Host:     "127.0.0.1",
		Port:     9002,
		Protocol: geeRpc.HTTP,
	}
	return c
}
