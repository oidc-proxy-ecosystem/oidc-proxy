package adapter

import (
	"github.com/oidc-proxy-ecosystem/oidc-proxy/adapter/command"
	"github.com/urfave/cli/v2"
)

type Adapter interface {
	Run(args []string) error
}

// New CLIアプリケーションのインスタンスメソッド
// github.com/urfave/cli/v2 を使用してCLI作成
func New(version, revision string) Adapter {
	app := cli.NewApp()
	app.Name = "openid connect proxy server"
	app.Version = version + " - " + revision
	app.Description = "openid connect proxy server"
	app.Commands = []*cli.Command{
		command.ProxyCommand,
		command.AppFileCommand,
	}
	return app
}
