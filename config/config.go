package config

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"

	"github.com/mcuadros/go-defaults"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/translate"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/validation"
	toml "github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// Config
type Config struct {
	Logging           Logging             `yaml:"logging" toml:"logging" json:"logging"`
	Servers           []*Servers          `yaml:"servers" toml:"servers" json:"servers" validate:"required=deep"`
	Port              int                 `yaml:"port" toml:"port" json:"port"`
	SslCertificate    string              `yaml:"ssl_certificate" toml:"ssl_certificate" json:"ssl_certificate"`
	SslCertificateKey string              `yaml:"ssl_certificate_key" toml:"ssl_certificate_key" json:"ssl_certificate_key"`
	Local             string              `yaml:"local" json:"local" toml:"local" default:"ja"`
	port              string              `yaml:"-" toml:"-" json:"-"`
	mapSrvConfig      map[string]*Servers `yaml:"-" toml:"-" json:"-"`
}

var _ Configuration = new(Config)

func (c *Config) GetPort() string {
	if c.port == "" {
		port := strconv.Itoa(c.Port)
		if port != "" && port[:1] != ":" {
			port = ":" + port
		}
		c.port = port
	}
	return c.port
}

func (c *Config) GetServerConfig(serverName string) *Servers {
	return c.mapSrvConfig[serverName]
}

func (c *Config) Output(filename string) error {
	typ := GetExtension(filename)
	var marshal func(interface{}) ([]byte, error)
	switch typ {
	case Yaml:
		marshal = yaml.Marshal
	case Toml:
		marshal = toml.Marshal
	case Json:
		marshal = json.Marshal
	}
	buf, err := marshal(c)
	if err != nil {
		return err
	}
	if typ == Json {
		var b bytes.Buffer
		if err := json.Indent(&b, buf, "", " "); err != nil {
			return err
		}
		buf = b.Bytes()
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(buf)
	return err
}

func (c *Config) Valid() error {
	var errs []error
	defaults.SetDefaults(c)
	valid = validation.New(validation.Translate(translate.LocalType(c.Local)))
	if err := c.autoSettings(); err != nil {
		errs = append(errs, err)
	}
	if err := valid.StructValidate(c); err != nil {
		if valErr, ok := err.(*validation.ValidateError); ok {
			errs = append(errs, valErr.Errors...)
		} else {
			errs = append(errs, err)
		}
	}
	if err := c.Logging.Valid(); err != nil {
		errs = append(errs, err)
	}
	for _, s := range c.Servers {
		if err := s.Valid(); err != nil {
			errs = append(errs, err)
		}
	}
	return newWrapError(errs...)
}

func (c *Config) autoSettings() error {
	if c.Port == 0 {
		port, err := findPort()
		if err != nil {
			return err
		}
		c.Port = port
	}
	return nil
}
