package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/config"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/routes"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/watch"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type Watch struct {
	mu                sync.RWMutex
	applicationConfig string
	Config            *config.Config
	MultiHost         routes.MultiHost
}

var _ watch.Watcher = &Watch{}

func New(applicationConfig string) (watch.Watcher, error) {
	var err error
	if !fileIsExists(applicationConfig) {
		return nil, watch.ErrFileNotFound
	}
	applicationConfig, err = filepath.Abs(applicationConfig)
	if err != nil {
		return nil, err
	}
	cm := &Watch{
		mu:                sync.RWMutex{},
		applicationConfig: applicationConfig,
	}
	return cm, nil
}

func (cm *Watch) Watch(watcher *fsnotify.Watcher) error {
	if err := watcher.Add(cm.applicationConfig); err != nil {
		return errors.Wrap(err, "configファイルを監視出来ません。")
	}
	return nil
}

func (cm *Watch) Load() error {
	appConf, err := config.New(cm.applicationConfig)
	cm.Config = &appConf
	multiHost := routes.NewMultiHostServe()
	for _, conf := range appConf.Servers {
		handler, err := routes.New(cm.GetConfiguration(conf.ServerName))
		if err != nil {
			return err
		}
		multiHost.Add(conf.GetHostname(), handler)
		logger.Log.Info(fmt.Sprintf("Server listening on %s", conf.GetHostname()))
	}
	cm.MultiHost = multiHost
	return err
}

func (cm *Watch) GetConfiguration(serverName string) config.GetConfiguration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return func() config.Servers {
		conf := *cm.Config.GetServerConfig(serverName)
		return conf
	}
}

func fileIsExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
