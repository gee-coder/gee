package main

import (
	"net/http"
	"time"

	"github.com/gee-coder/gee"
	"github.com/gee-coder/gee/gateway"
	"github.com/gee-coder/gee/register"
)

func main() {
	engine := gee.Default()
	engine.OpenGateWay = true
	var configs []gateway.GWConfig
	configs = append(configs, gateway.GWConfig{
		Name: "order",
		Path: "/order/**",
		Header: func(req *http.Request) {
			req.Header.Set("my", "geecoder")
		},
		ServiceName: "orderCenter",
	}, gateway.GWConfig{
		Name: "goods",
		Path: "/goods/**",
		Header: func(req *http.Request) {
			req.Header.Set("my", "geecoder")
		},
		ServiceName: "goodsCenter",
	})
	engine.SetGatewayConfig(configs)
	engine.RegisterType = "etcd"
	engine.RegisterOption = register.Option{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}
	engine.Run(":80")
}
