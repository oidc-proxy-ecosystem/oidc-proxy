package routes_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/routes"
	"github.com/stretchr/testify/assert"
)

func TestA(t *testing.T) {
	u, _ := url.Parse("https://kubernetes-dashboard.com")
	r, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	hostPort := routes.GetHostName(r)
	assert.Equal(t, "kubernetes-dashboard.com:443", hostPort)
}

func TestB(t *testing.T) {
	u, _ := url.Parse("https://kubernetes-dashboard.com:8443")
	r, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	hostPort := routes.GetHostName(r)
	assert.Equal(t, "kubernetes-dashboard.com:8443", hostPort)
}

func TestC(t *testing.T) {
	u, _ := url.Parse("http://kubernetes-dashboard.com")
	r, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	hostPort := routes.GetHostName(r)
	assert.Equal(t, "kubernetes-dashboard.com:80", hostPort)
}

func TestD(t *testing.T) {
	u, _ := url.Parse("http://kubernetes-dashboard.com:8080")
	r, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	hostPort := routes.GetHostName(r)
	assert.Equal(t, "kubernetes-dashboard.com:8080", hostPort)
}
