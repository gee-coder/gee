package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gee-coder/gee"
	geeRpc "github.com/gee-coder/gee/rpc"
	"github.com/gee-coder/goodscenter/api"
	"github.com/gee-coder/goodscenter/model"
	"github.com/gee-coder/goodscenter/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	engine := gee.Default()
	client := geeRpc.NewHttpClient()
	client.RegisterHttpService("goodsService", &service.GoodsService{})
	session := client.Session()

	group := engine.Group("orders")
	group.Get("/find", func(ctx *gee.Context) {
		params := make(map[string]any)
		params["id"] = 1000
		params["name"] = "mi"
		body, err := session.Do("goodsService", "Find").(*service.GoodsService).Find(params)
		if err != nil {
			panic(err)
		}
		log.Printf(string(body))
		v := &model.Result{}
		err = json.Unmarshal(body, v)
		if err != nil {
			panic(err)
		}
		err = ctx.JSON(http.StatusOK, v)
		if err != nil {
			panic(err)
		}
	})

	group.Get("/findGrpc", func(ctx *gee.Context) {
		// 查询商品
		var serviceHost = "127.0.0.1:9111"
		conn, err := grpc.NewClient(serviceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println(err)
		}
		defer conn.Close()
		client := api.NewGoodsApiClient(conn)
		rsp, err := client.Find(context.TODO(), &api.GoodsRequest{})
		if err != nil {
			fmt.Println(err)
		}
		ctx.JSON(http.StatusOK, rsp)
	})

	engine.Run(":9003")
}
