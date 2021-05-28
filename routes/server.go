package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/config"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MultiHost interface {
	http.Handler
	Add(virtualServerName string, h Handler)
}

type multiHost struct {
	multiHost map[string]Handler
	mux       *http.ServeMux
}

func NewMultiHostServe() MultiHost {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return &multiHost{
		multiHost: make(map[string]Handler),
		mux:       mux,
	}
}

func (m *multiHost) Add(virtualServerName string, h Handler) {
	m.multiHost[virtualServerName] = h
}

func (m *multiHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := m.multiHost[GetHostName(r)]; handler != nil {
		handler.ServeHTTP(w, r)
	} else if h, patten := m.mux.Handler(r); h != nil && patten != "" {
		logger.Log.Info(r.URL.String())
		h.ServeHTTP(w, r)
	} else {
		err := fmt.Errorf("%s: %s", GetHostName(r), http.StatusText(http.StatusNotFound))
		logger.Log.Error(err)
		msg := responseErrorPage(http.StatusNotFound, err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(msg))
	}
}

func New(configuration config.GetConfiguration) (Handler, error) {
	conf := configuration()
	router, err := new(conf)
	if err != nil {
		return nil, err
	}
	host := conf.GetHostname()
	for _, location := range conf.Locations {
		proxypasses := strings.Split(location.ProxyPass, ",")
		registry, err := NewRegistry(proxypasses)
		if err != nil {
			return nil, err
		}
		for _, path := range location.Urls {
			router.Proxy(path.Path, registry, host, path.Type, path.Token, location.IsProxySSLVerify())
		}
	}
	router.Login(conf.Login)
	router.Callback(conf.Callback)
	router.Logout(conf.Logout)
	return router, nil
}

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
