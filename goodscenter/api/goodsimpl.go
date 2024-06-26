package api

import (
	"context"
)

type GoodsRpcService struct {
}

func (GoodsRpcService) Find(context.Context, *GoodsRequest) (*GoodsResponse, error) {
	goods := &Goods{Id: 1000, Name: "商品中心9002商品,grpc提供"}
	res := &GoodsResponse{
		Code: 200,
		Msg:  "success",
		Data: goods,
	}
	return res, nil
}
func (GoodsRpcService) mustEmbedUnimplementedGoodsApiServer() {}
