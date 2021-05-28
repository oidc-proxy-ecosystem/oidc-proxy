package config

import (
	"io"
	"os"

	"github.com/mcuadros/go-defaults"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/validation"
)

// Logging
type Logging struct {
	Level      string    `yaml:"level" toml:"level" json:"level" default:"info"`
	FileName   string    `yaml:"filename" toml:"filename" json:"filename"`
	LogFormat  string    `yaml:"logformat" toml:"logformat" json:"logformat" default:"standard"`
	TimeFormat string    `yaml:"timeformat" toml:"timeformat" json:"timeformat" default:"datetime"`
	writer     io.Writer `yaml:"-"`
}

var _ Configuration = new(Logging)

func (l *Logging) GetLogger() logger.ILogger {
	writer := []io.Writer{}
	if l.FileName != "" {
		if logfile, err := os.OpenFile(l.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err == nil {
			writer = append(writer, logfile)
		} else {
			logger.Log.Error(err.Error())
		}
	} else {
		writer = append(writer, os.Stdout)
	}
	l.writer = io.MultiWriter(writer...)
	return logger.New(l.writer, logger.Convert(l.Level), logger.ConvertLogFmt(l.LogFormat), logger.ConvertTimeFmt(l.TimeFormat))
}

func (l *Logging) Valid() error {
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
	return newWrapError(errs...)
}

func (l *Logging) autoSettings() error {
	return nil
}
