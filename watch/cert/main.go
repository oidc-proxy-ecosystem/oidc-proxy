package cert

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"sync"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/watch"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type Watch struct {
	mu       sync.RWMutex
	certFile string
	keyFile  string
	keyPair  *tls.Certificate
}

var _ watch.Watcher = &Watch{}

func New(certFile, keyFile string) (watch.Watcher, error) {
	var err error
	if !fileIsExists(certFile) {
		return nil, watch.ErrFileNotFound
	}
	if !fileIsExists(keyFile) {
		return nil, watch.ErrFileNotFound
	}
	certFile, err = filepath.Abs(certFile)
	if err != nil {
		return nil, err
	}
	keyFile, err = filepath.Abs(keyFile)
	if err != nil {
		return nil, err
	}
	cm := &Watch{
		mu:       sync.RWMutex{},
		certFile: certFile,
		keyFile:  keyFile,
	}
	return cm, nil
}

func (cm *Watch) Watch(watcher *fsnotify.Watcher) error {
	if err := watcher.Add(cm.certFile); err != nil {
		return errors.Wrap(err, "certファイルを監視出来ません。")
	}
	if err := watcher.Add(cm.keyFile); err != nil {
		return errors.Wrap(err, "keyファイルを監視出来ません。")
	}
	return nil
}

func (cm *Watch) Load() error {
	keyPair, err := tls.LoadX509KeyPair(cm.certFile, cm.keyFile)
	if err == nil {
		cm.mu.Lock()
		defer cm.mu.Unlock()
		cm.keyPair = &keyPair
	}
	return err
}

func (cm *Watch) GetCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.keyPair, nil
}

func fileIsExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
