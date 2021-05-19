package app

import (
	"os/exec"
	"sync"

	"github.com/hashicorp/go-plugin"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/store"
)

type Dispose struct {
	Store  *store.SessionStore
	Plugin *plugin.Client
	Cmd    *exec.Cmd
}

func (d *Dispose) Close() error {
	var err error
	if d.Store != nil {
		err = d.Store.Close()
	}
	if d.Plugin != nil {
		d.Plugin.Kill()
	}
	return err
}

type StoreMap struct {
	mu    sync.Mutex
	store map[string]*Dispose
}

func (s *StoreMap) Add(name string, dispose *Dispose) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[name] = dispose
}

func (s *StoreMap) Dispose(name string) *Dispose {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.store[name]
}

func (s *StoreMap) Store(name string) *store.SessionStore {
	return s.Dispose(name).Store
}

var Store = StoreMap{
	store: map[string]*Dispose{},
	mu:    sync.Mutex{},
}
