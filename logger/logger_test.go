package logger_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "logger debug",
			fn: func(t *testing.T) {
				buffer := new(bytes.Buffer)
				logging := logger.New(buffer, logger.Debug, logger.FormatStandard, logger.FormatDatetime)
				now := time.Now().Format(logger.FormatDatetime.String())
				logging.Debug("debug test")
				assert.Equal(t, buffer.String(), fmt.Sprintf("time:%s\tlevel:[DEBUG]\tfilename:logger_test.go:24\tmessage:debug test\n", now))
				buffer.Reset()
				logging.Info("info test")
				assert.Equal(t, buffer.String(), fmt.Sprintf("time:%s\tlevel:[INFO]\tfilename:logger_test.go:27\tmessage:info test\n", now))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
