package config

import (
	"github.com/mcuadros/go-defaults"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/validation"
)

// Locations
type Locations struct {
	ProxyPass      string `yaml:"proxy_pass" toml:"proxy_pass" json:"proxy_pass" validate:"required"`
	ProxySSLVerify string `yaml:"proxy_ssl_verify" toml:"proxy_ssl_verify" json:"proxy_ssl_verify" default:"false"`
	Urls           []Urls `yaml:"urls" toml:"urls" json:"urls" validate:"required"`
}

var _ Configuration = new(Locations)

func (l *Locations) IsProxySSLVerify() bool {
	return l.ProxySSLVerify == "on"
}

func (l *Locations) Valid() error {
	var errs []error
	defaults.SetDefaults(l)
	if err := l.autoSettings(); err != nil {
		errs = append(errs, err)
	}
	if err := valid.StructValidate(l); err != nil {
		if valErr, ok := err.(*validation.ValidateError); ok {
			errs = append(errs, valErr.Errors...)
		} else {
			errs = append(errs, err)
		}
	}
	for _, u := range l.Urls {
		if err := u.Valid(); err != nil {
			errs = append(errs, err)
		}
	}
	if err := l.autoSettings(); err != nil {
		errs = append(errs, err)
	}
	return newWrapError(errs...)
}

func (l *Locations) autoSettings() error {
	return nil
}
