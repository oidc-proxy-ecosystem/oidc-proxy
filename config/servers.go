package config

import (
	"context"
	"os/exec"

	hplugin "github.com/hashicorp/go-plugin"
	"github.com/mcuadros/go-defaults"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/app"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/plugin"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/session"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/store"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/validation"
	"github.com/prometheus/common/log"
)

// Servers
type Servers struct {
	Oidc       Oidc        `yaml:"oidc" toml:"oidc" json:"oidc" validate:"required"`
	Locations  []Locations `yaml:"locations" toml:"locations" json:"locations" validate:"required"`
	Logging    Logging     `yaml:"logging" toml:"logging" json:"logging"`
	Session    Session     `yaml:"session" toml:"session" json:"session"`
	CookieName string      `yaml:"cookie_name" toml:"cookie_name" json:"cookie_name" default:"session"`
	ServerName string      `yaml:"server_name" toml:"server_name" json:"server_name" validate:"required"`
	Login      string      `yaml:"login" toml:"login" json:"login" default:"/oauth2/login"`
	Callback   string      `yaml:"callback" toml:"callback" json:"callback" default:"/oauth2/callback"`
	Logout     string      `yaml:"logout" toml:"logout" json:"logout" default:"/oauth2/logout"`
	Redirect   bool        `yaml:"redirect" toml:"redirect" json:"redirect" default:"true"`
}

var _ Configuration = new(Servers)

func (s *Servers) newSessionPluginClient(client *hplugin.Client) session.Session {
	var storage session.Session
	rpcClient, err := client.Client()
	if err != nil {
		log.Error(err)
	}
	if rpcClient != nil {
		raw, err := rpcClient.Dispense("session")
		if err != nil {
			log.Error(err)
		}
		if raw != nil {
			if val, ok := raw.(*plugin.GRPCSessionClient); ok {
				val.PluginClient = client
				storage = val
			}
		}
	}
	_ = storage.Init(context.TODO(), s.Session.Args)
	return storage
}

func (s *Servers) init() {
	var storage session.Session
	var sessionStore *store.SessionStore
	var sessionPlugin *hplugin.Client
	if s.Session.Plugin {
		cmd := exec.Command(s.Session.GetSessionPlugin())
		dispose := app.Store.Dispose(s.ServerName)
		if dispose != nil && dispose.Plugin != nil {
			dispose.Close()
		}
		sessionPlugin = plugin.NewClient(cmd, s.Logging.Level, s.Logging.writer)
		storage = s.newSessionPluginClient(sessionPlugin)
	} else {
		storage = session.NewLocalMemory()
		storage.Init(context.Background(), s.Session.Args)
	}
	codecs := s.Session.GetCodecs()
	sessionStore = store.NewStore(storage, codecs...)
	app.Store.Add(s.ServerName, &app.Dispose{
		Store:  sessionStore,
		Plugin: sessionPlugin,
	})
}

func (s *Servers) Valid() error {
	var errs []error
	defaults.SetDefaults(s)
	if err := s.autoSettings(); err != nil {
		errs = append(errs, err)
	}
	if err := valid.StructValidate(s); err != nil {
		if valErr, ok := err.(*validation.ValidateError); ok {
			errs = append(errs, valErr.Errors...)
		} else {
			errs = append(errs, err)
		}
	}
	if err := s.Oidc.Valid(); err != nil {
		errs = append(errs, err)
	}
	for _, location := range s.Locations {
		if err := location.Valid(); err != nil {
			errs = append(errs, err)
		}
	}
	if err := s.Logging.Valid(); err != nil {
		errs = append(errs, err)
	}
	if err := s.Session.Valid(); err != nil {
		errs = append(errs, err)
	}
	if err := s.autoSettings(); err != nil {
		errs = append(errs, err)
	}
	return newWrapError(errs...)
}

func (s *Servers) autoSettings() error {
	return nil
}

func (s *Servers) GetHostname() string {
	return s.ServerName
}
