package validation

import "github.com/oidc-proxy-ecosystem/oidc-proxy/translate"

type Options struct {
	Translate map[string]string
}

type Builder struct {
	Options *Options
}

type Option func(builder *Builder)

func Translate(typ translate.LocalType) Option {
	return func(builder *Builder) {
		builder.Options.Translate = translate.New(typ)
	}
}

func New(options ...Option) Validate {
	b := &Builder{
		Options: &Options{
			Translate: translate.New(translate.Japan),
		},
	}
	for _, opts := range options {
		opts(b)
	}
	return &validateImpl{
		builder: b,
	}
}
