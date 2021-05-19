package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/config"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
)

type MultiHost map[string]Handler

func GetHostName(r *http.Request) string {
	hostPortSplit := strings.Split(r.Host, ":")
	if len(hostPortSplit) > 1 {
		return r.Host
	}
	port := ""
	if r.URL.Scheme == "https" {
		port = "443"
	} else {
		port = "80"
	}
	return r.Host + ":" + port
}

func (m MultiHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := m[GetHostName(r)]; handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		err := fmt.Errorf("%s: %s", GetHostName(r), http.StatusText(http.StatusNotFound))
		errorResponse(logger.Log)(w, r, err)
	}
}

func New(configuration config.GetConfiguration) (Handler, error) {
	conf := configuration()
	router := new(conf)
	host := conf.GetHostname()
	for _, location := range conf.Locations {
		u, err := url.Parse(location.ProxyPass)
		if err != nil {
			return nil, err
		}
		for _, path := range location.Urls {
			router.Proxy(path.Path, u, host, path.Type, path.Token, location.IsProxySSLVerify())
		}
	}
	router.Login(conf.Login)
	router.Callback(conf.Callback)
	router.Logout(conf.Logout)
	return router, nil
}
