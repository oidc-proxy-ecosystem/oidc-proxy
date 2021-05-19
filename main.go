package main

import (
	"os"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/adapter"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
)

func main() {
	app := adapter.New(Version, Revision)
	if err := app.Run(os.Args); err != nil {
		logger.Log.Error(err)
	}
}
