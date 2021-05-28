package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/validation"
	toml "github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

var ErrInvalidSpecification = errors.New("specification must be a struct pointer")
var valid = validation.New()

type Configuration interface {
	Valid() error
	autoSettings() error
}

type GetConfiguration func() Servers

func ReadConfig(filename string) (Config, error) {
	var conf Config
	buf, err := os.ReadFile(filename)
	if err != nil {
		return conf, err
	}
	expand := os.ExpandEnv(string(buf))
	ext := GetExtension(filename)
	var unmarshal func([]byte, interface{}) error
	switch ext {
	case Yaml:
		unmarshal = yaml.Unmarshal
	case Toml:
		unmarshal = toml.Unmarshal
	case Json:
		unmarshal = json.Unmarshal
	default:
		unmarshal = yaml.Unmarshal
	}
	err = unmarshal([]byte(expand), &conf)
	if err != nil {
		return conf, err
	}
	err = conf.Valid()
	return conf, err
}

type Extension int

const (
	Yaml Extension = iota
	Toml
	Json
)

func GetExtension(filename string) Extension {
	ext := filepath.Ext(filename)
	switch ext {
	case ".yaml", ".yml":
		return Yaml
	case ".toml":
		return Toml
	case ".json":
		return Json
	default:
		return Yaml
	}
}

func New(filename string) (Config, error) {
	conf, err := ReadConfig(filename)
	if err != nil {
		return conf, err
	}
	conf.mapSrvConfig = map[string]*Servers{}
	for _, s := range conf.Servers {
		server := s
		if err := server.Valid(); err != nil {
			return conf, err
		}
		server.init()
		conf.mapSrvConfig[s.ServerName] = server
	}
	return conf, nil
}
