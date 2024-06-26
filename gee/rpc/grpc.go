package rpc

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// listen, _ := net.Listen("tcp", ":9111")
//	server := grpc.NewServer()
//	api.RegisterGoodsApiServer(server, &api.GoodsRpcService{})
//	err := server.Serve(listen)

type GeeGrpcServer struct {
	listen   net.Listener
	g        *grpc.Server
	register []func(g *grpc.Server)
	ops      []grpc.ServerOption
}

func NewGrpcServer(addr string, ops ...MsGrpcOption) (*GeeGrpcServer, error) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	ms := &GeeGrpcServer{}
	ms.listen = listen
	for _, v := range ops {
		v.Apply(ms)
	}
	server := grpc.NewServer(ms.ops...)
	ms.g = server
	return ms, nil
}

func (s *GeeGrpcServer) Run() error {
	for _, f := range s.register {
		f(s.g)
	}
	return s.g.Serve(s.listen)
}

func (s *GeeGrpcServer) Stop() {
	s.g.Stop()
}

func (s *GeeGrpcServer) Register(f func(g *grpc.Server)) {
	s.register = append(s.register, f)
}

type MsGrpcOption interface {
	Apply(s *GeeGrpcServer)
}

type DefaultMsGrpcOption struct {
	f func(s *GeeGrpcServer)
}

func (d *DefaultMsGrpcOption) Apply(s *GeeGrpcServer) {
	d.f(s)
}

func WithGrpcOptions(ops ...grpc.ServerOption) MsGrpcOption {
	return &DefaultMsGrpcOption{
		f: func(s *GeeGrpcServer) {
			s.ops = append(s.ops, ops...)
		},
	}
}

type GeeGrpcClient struct {
	Conn *grpc.ClientConn
}

func NewGrpcClient(config *GeeGrpcClientConfig) (*GeeGrpcClient, error) {
	var ctx = context.Background()
	var dialOptions = config.dialOptions

	if config.Block {
		// 阻塞
		if config.DialTimeout > time.Duration(0) {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, config.DialTimeout)
			defer cancel()
		}
		dialOptions = append(dialOptions, grpc.WithBlock())
	}
	if config.KeepAlive != nil {
		dialOptions = append(dialOptions, grpc.WithKeepaliveParams(*config.KeepAlive))
	}
	conn, err := grpc.DialContext(ctx, config.Address, dialOptions...)
	if err != nil {
		return nil, err
	}
	return &GeeGrpcClient{
		Conn: conn,
	}, nil
}

type GeeGrpcClientConfig struct {
	Address     string
	Block       bool
	DialTimeout time.Duration
	ReadTimeout time.Duration
	Direct      bool
	KeepAlive   *keepalive.ClientParameters
	dialOptions []grpc.DialOption
}

func DefaultGrpcClientConfig() *GeeGrpcClientConfig {
	return &GeeGrpcClientConfig{
		dialOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
		DialTimeout: time.Second * 3,
		ReadTimeout: time.Second * 2,
		Block:       true,
	}
}
