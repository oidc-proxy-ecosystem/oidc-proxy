package plugin

import (
	hplugin "github.com/hashicorp/go-plugin"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/plugin"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/session"
	"google.golang.org/grpc"
)

const SeesionPluginName = "session"

var Handshake = hplugin.HandshakeConfig{
	MagicCookieKey:   "SESSION_PLUGIN",
	MagicCookieValue: "m9erzlkcuac9gy4a2szc19j7xjleo4s4epwiio9opv8tjv9sid0qetl7cjo6ulkiskorqyg26pcsfyf979pgn28s5a7byfbq0n66",
}

type GRPCSessionFunc func() session.Session

type ServerOpts struct {
	GRPCSessionFunc GRPCSessionFunc
	TestConfig      *hplugin.ServeTestConfig
}

func Sever(opts *ServerOpts) {
	provider := opts.GRPCSessionFunc()
	hplugin.Serve(&hplugin.ServeConfig{
		HandshakeConfig: Handshake,
		VersionedPlugins: map[int]hplugin.PluginSet{
			1: {
				SeesionPluginName: &plugin.GRPCSessionPlugin{
					Impl: provider,
				},
			},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			return grpc.NewServer(opts...)
		},
		Test: opts.TestConfig,
	})
}
