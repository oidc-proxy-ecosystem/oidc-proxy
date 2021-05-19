package plugin

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/internal/proto"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/session"
	"google.golang.org/grpc"
)

// rpcのサポートをする場合
// type CachePlugin struct {
// 	Impl cache.Cache
// }

// func (p *CachePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
// 	return nil, nil
// }

type GRPCSessionPlugin struct {
	plugin.Plugin
	Impl session.Session
}

func (p *GRPCSessionPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterSessionServer(s, &GRPCSessionServer{
		Impl: p.Impl,
	})
	return nil
}
func (p *GRPCSessionPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCSessionClient{
		client: proto.NewSessionClient(c),
	}, nil
}
