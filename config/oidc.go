package config

import (
	"github.com/mcuadros/go-defaults"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/validation"
	"golang.org/x/oauth2"
)

// Oidc
type Oidc struct {
	Scopes       []string `yaml:"scopes" toml:"scopes" json:"scopes" default:"[email,openid,offline_access,profile]"`
	Provider     string   `yaml:"provider" toml:"provider" json:"provider" validate:"required"`
	ClientId     string   `yaml:"client_id" toml:"client_id" json:"client_id" validate:"required"`
	ClientSecret string   `yaml:"client_secret" toml:"client_secret" json:"client_secret" validate:"required"`
	RedirectUrl  string   `yaml:"redirect_url" toml:"redirect_url" json:"redirect_url" validate:"required"`
	Logout       string   `yaml:"logout" toml:"logout" json:"logout"`
	Audiences    []string `yaml:"audiences" toml:"audiences" json:"audiences"`
}

var _ Configuration = new(Oidc)

func (o *Oidc) SetValues() []oauth2.AuthCodeOption {
	var authCodeOptions []oauth2.AuthCodeOption
	var audiences Audiences
	for _, audience := range o.Audiences {
		audiences = append(audiences, Audience(audience))
	}
	authCodeOptions = append(authCodeOptions, audiences.SetValue()...)
	return authCodeOptions
}

func (o *Oidc) Valid() error {
	var errs []error
	defaults.SetDefaults(o)

	if err := valid.StructValidate(o); err != nil {
		if valErr, ok := err.(*validation.ValidateError); ok {
			errs = append(errs, valErr.Errors...)
		} else {
			errs = append(errs, err)
		}
	}
	if err := o.autoSettings(); err != nil {
		errs = append(errs, err)
	}
	return newWrapError(errs...)
}

func (o *Oidc) autoSettings() error {
	return nil
}
