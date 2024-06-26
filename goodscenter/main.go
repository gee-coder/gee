package main

import (
	"log"
	"net"
	"net/http"

	"github.com/gee-coder/gee"
	"github.com/gee-coder/goodscenter/api"
	"github.com/gee-coder/goodscenter/model"
	"google.golang.org/grpc"
)

func main() {

	engine := gee.Default()
	group := engine.Group("goods")
	group.Get("/find", func(ctx *gee.Context) {
		goods := &model.Goods{Id: 1000, Name: "9002的商品"}
		ctx.JSON(http.StatusOK, &model.Result{Code: 200, Msg: "success", Data: goods})
	})

	listen, _ := net.Listen("tcp", ":9111")
	server := grpc.NewServer()
	api.RegisterGoodsApiServer(server, &api.GoodsRpcService{})
	err := server.Serve(listen)
	log.Println(err)

	engine.Run(":9002")
}
