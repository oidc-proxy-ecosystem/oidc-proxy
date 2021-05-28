package config

import (
	"github.com/mcuadros/go-defaults"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/validation"
)

// Urls
type Urls struct {
	Path  string `yaml:"path" toml:"path" json:"path" validate:"required"`
	Token string `yaml:"token" toml:"token" json:"token" validate:"required" default"id_token"`
	Type  string `yaml:"type" toml:"type" json:"type" validate:"required" default:"bearer"`
}

var _ Configuration = new(Urls)

func (u *Urls) Valid() error {
	var errs []error
	defaults.SetDefaults(u)
	if err := u.autoSettings(); err != nil {
		errs = append(errs, err)
	}
	if err := valid.StructValidate(u); err != nil {
		if valErr, ok := err.(*validation.ValidateError); ok {
			errs = append(errs, valErr.Errors...)
		} else {
			errs = append(errs, err)
		}
	}
	if err := u.autoSettings(); err != nil {
		errs = append(errs, err)
	}
	return newWrapError(errs...)
}

func (u *Urls) autoSettings() error {
	return nil
}
