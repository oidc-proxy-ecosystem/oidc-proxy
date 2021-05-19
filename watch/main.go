package watch

import (
	"fmt"
	"sync"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

var (
	ErrFileNotFound = errors.New("No such file or directory")
)

type Watcher interface {
	Watch(watcher *fsnotify.Watcher) error
	Load() error
}

type Watch struct {
	mu       sync.RWMutex
	watcher  *fsnotify.Watcher
	watching chan bool
	log      logger.ILogger
	Watching Watcher
}

func New(log logger.ILogger) (*Watch, error) {
	if watcher, err := fsnotify.NewWatcher(); err != nil {
		return nil, errors.Wrap(err, "")
	} else {
		return &Watch{
			log:     log,
			watcher: watcher,
		}, nil
	}
}

func (w *Watch) Watch() error {
	if err := w.Watching.Watch(w.watcher); err != nil {
		return errors.Wrap(err, "")
	}
	if err := w.Load(); err != nil {
		w.log.Error("ファイルロードエラー: %v", err)
	}
	w.log.Info("監視を開始します。")
	w.watching = make(chan bool)
	go w.Run()
	return nil
}

func (w *Watch) Load() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Watching.Load()
}

func (w *Watch) Run() {
loop:
	for {
		select {
		case <-w.watching:
			break loop
		case event := <-w.watcher.Events:
			w.log.Info(fmt.Sprintf("監視イベント: %v", event))
			if err := w.Load(); err != nil {
				w.log.Error(fmt.Sprintf("ファイルロードエラー: %v", err))
			}
		case err := <-w.watcher.Errors:
			w.log.Error(fmt.Sprintf("監視ファイルにエラーが発生しました。: %v", err))
		}
	}
	w.log.Info("ファイル監視を終了します。")
	w.watcher.Close()
}

func (w *Watch) Stop() {
	w.watching <- false
}
