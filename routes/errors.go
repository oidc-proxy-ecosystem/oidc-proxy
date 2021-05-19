package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
)

const (
	httpStatusClientClosedRequest = 499
)

func errorResponse(log logger.ILogger) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		status := http.StatusBadGateway
		switch err {
		case context.Canceled:
			status = httpStatusClientClosedRequest
		case io.ErrUnexpectedEOF:
			status = httpStatusClientClosedRequest
		default:
			log.Error(fmt.Sprintf("http: proxy error : %v", err))
		}
		w.WriteHeader(status)
	}
}

type errorBody struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func UnAuthorizedResponse(w http.ResponseWriter, location string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if location != "" {
		w.Header().Set("Location", location)
	}
	w.WriteHeader(http.StatusUnauthorized)
	errBody := errorBody{
		StatusCode: http.StatusUnauthorized,
		Message:    http.StatusText(http.StatusUnauthorized),
	}
	buf, err := json.Marshal(&errBody)
	if err == nil {
		w.Write(buf)
	}
}

func responseError(log logger.ILogger, w http.ResponseWriter, err string, code int) {
	log.Critical(err)
	http.Error(w, err, code)
}
