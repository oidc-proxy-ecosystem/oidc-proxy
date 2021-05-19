package plugin

import (
	"io"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

var VersionedPlugins = map[int]plugin.PluginSet{
	1: {
		"session": &GRPCSessionPlugin{},
	},
}

var Handshake = plugin.HandshakeConfig{
	MagicCookieKey:   "SESSION_PLUGIN",
	MagicCookieValue: "m9erzlkcuac9gy4a2szc19j7xjleo4s4epwiio9opv8tjv9sid0qetl7cjo6ulkiskorqyg26pcsfyf979pgn28s5a7byfbq0n66",
}

func NewClient(cmd *exec.Cmd, loglevel string, writer io.Writer) *plugin.Client {
	return plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  Handshake,
		VersionedPlugins: VersionedPlugins,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger: hclog.New(&hclog.LoggerOptions{
			Level:      hclog.LevelFromString(loglevel),
			Output:     writer,
			TimeFormat: "2006/01/02 - 15:04:05",
			Name:       "session",
		}),
		Managed: true,
	})
}
