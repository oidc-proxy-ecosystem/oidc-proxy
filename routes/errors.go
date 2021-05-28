package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
)

const (
	httpStatusClientClosedRequest = 499
)

var (
	pages = make(map[int]string)
)

func MustRegistryPages(code int, pageHtml string) {
	pages[code] = strings.TrimSpace(pageHtml)
}

func responseErrorPage(code int, respErr error) string {
	if val, ok := pages[code]; ok {
		mp := map[string]interface{}{
			"Title": http.StatusText(code),
			"Err":   respErr.Error(),
		}
		tmpl, err := template.New("html").Parse(val)
		if err != nil {
			return ""
		}
		buf := bytes.NewBufferString("")
		if err := tmpl.Execute(buf, mp); err != nil {
			return ""
		}
		return buf.String()
	}
	return respErr.Error()
}

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
		switch err.(type) {
		case *ErrNotFoundPage:
			msg := responseErrorPage(http.StatusNotFound, err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(msg))
		default:
			w.WriteHeader(status)
		}
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
