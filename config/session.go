package config

import (
	"path/filepath"

	"github.com/mcuadros/go-defaults"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/validation"
)

// Session
type Session struct {
	Name   string                 `yaml:"name" toml:"name" json:"name"`
	Plugin bool                   `yaml:"plugin" toml:"plugin" json:"plugin"`
	Codecs []string               `yaml:"codecs" toml:"codecs" json:"codecs" validate:"required"`
	Args   map[string]interface{} `yaml:"args" toml:"args" json:"args"`
}

var _ Configuration = &Session{}

const defaultSessionPlugin = "memory"

func (s *Session) GetSessionPlugin() string {
	const pluginDir = "oidc-plugin"
	if s.Name != "" {
		return filepath.Join("./", pluginDir, s.Name)
	}
	return filepath.Join("./", pluginDir, defaultSessionPlugin)
}

func (s *Session) GetCodecs() [][]byte {
	var codecs [][]byte
	for _, codec := range s.Codecs {
		codecs = append(codecs, []byte(codec))
	}
	return codecs
}

func (s *Session) Valid() error {
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
	if err := s.autoSettings(); err != nil {
		errs = append(errs, err)
	}
	return newWrapError(errs...)
}

func (s *Session) autoSettings() error {
	return nil
}
