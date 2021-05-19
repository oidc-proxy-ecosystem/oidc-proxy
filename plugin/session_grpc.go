package plugin

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/go-plugin"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/internal/proto"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/session"
	"google.golang.org/grpc"
)

type GRPCSessionServer struct {
	Impl session.Session
	proto.UnimplementedSessionServer
}

var _ proto.SessionServer = &GRPCSessionServer{}

func (c *GRPCSessionServer) Init(ctx context.Context, r *proto.SettingRequest) (*proto.Empty, error) {
	var mp map[string]interface{}
	json.Unmarshal(r.Config, &mp)
	c.Impl.Init(ctx, mp)
	return &proto.Empty{}, nil
}
func (c *GRPCSessionServer) Get(ctx context.Context, r *proto.GetRequest) (*proto.GetResponse, error) {
	value, err := c.Impl.Get(ctx, r.Key)
	return &proto.GetResponse{
		Value: value,
	}, err
}
func (c *GRPCSessionServer) Put(ctx context.Context, r *proto.PutRequest) (*proto.Empty, error) {
	err := c.Impl.Put(ctx, r.Key, r.Value)
	return &proto.Empty{}, err
}
func (c *GRPCSessionServer) Delete(ctx context.Context, r *proto.DeleteRequest) (*proto.Empty, error) {
	err := c.Impl.Delete(ctx, r.Key)
	return &proto.Empty{}, err
}
func (c *GRPCSessionServer) Close(ctx context.Context, e *proto.Empty) (*proto.Empty, error) {
	err := c.Impl.Close(ctx)
	return &proto.Empty{}, err
}

type GRPCSessionClient struct {
	PluginClient *plugin.Client
	TestServer   *grpc.Server
	client       proto.SessionClient
}

var _ session.Session = &GRPCSessionClient{}

func (p *GRPCSessionClient) Get(ctx context.Context, key string) (string, error) {
	r := &proto.GetRequest{
		Key: key,
	}
	res, err := p.client.Get(ctx, r)
	if err != nil {
		return "", err
	}
	return res.Value, nil
}
func (p *GRPCSessionClient) Put(ctx context.Context, key string, value string) error {
	r := &proto.PutRequest{
		Key:   key,
		Value: value,
	}
	_, err := p.client.Put(ctx, r)
	if err != nil {
		return err
	}
	return nil
}
func (p *GRPCSessionClient) Delete(ctx context.Context, key string) error {
	r := &proto.DeleteRequest{
		Key: key,
	}
	_, err := p.client.Delete(ctx, r)
	if err != nil {
		return err
	}
	return nil
}
func (p *GRPCSessionClient) Init(ctx context.Context, setting map[string]interface{}) error {
	buf, _ := json.Marshal(setting)
	r := &proto.SettingRequest{
		Config: buf,
	}
	_, err := p.client.Init(ctx, r)
	return err
}
func (p *GRPCSessionClient) Close(ctx context.Context) error {
	if p.PluginClient == nil {
		return nil
	}
	_, err := p.client.Close(ctx, &proto.Empty{})
	return err
}
