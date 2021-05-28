package routes

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/app"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/auth"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/config"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/errors"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
	"golang.org/x/oauth2"
)

type handler struct {
	conf          config.Servers
	mux           *http.ServeMux
	log           logger.ILogger
	authenticator *auth.Authenticator
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = logger.NewContext(ctx, h.log)
	*r = *r.WithContext(ctx)
	h.mux.ServeHTTP(w, r)
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	c := NewClient(r.Context())
	ctx = context.WithValue(ctx, oauth2.HTTPClient, c)

	conf := h.conf
	if r.Method != http.MethodGet {
		return
	}
	// Generate random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		responseError(h.log, w, err.Error(), http.StatusInternalServerError)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	session, err := app.Store.Store(conf.ServerName).Get(r, conf.CookieName)
	if err != nil {
		responseError(h.log, w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["state"] = state
	err = session.Save(r, w)
	if err != nil {
		responseError(h.log, w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, h.authenticator.Config.AuthCodeURL(state, conf.Oidc.SetValues()...), http.StatusTemporaryRedirect)
}

func (h *handler) Login(pattern string) {
	h.mux.Handle(pattern, genInstrumentChain("login", h.login))
}

func (h *handler) callback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	c := NewClient(r.Context())
	ctx = context.WithValue(ctx, oauth2.HTTPClient, c)

	conf := h.conf
	if r.Method != http.MethodGet {
		return
	}
	session, err := app.Store.Store(conf.ServerName).Get(r, conf.CookieName)
	if err != nil {
		responseError(h.log, w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("state") != session.Values["state"] {
		h.log.Debug(fmt.Sprintf("request_state:%s", r.URL.Query().Get("state")))
		h.log.Debug(fmt.Sprintf("session_state:%s", session.Values["state"]))
		http.Redirect(w, r, conf.Login, http.StatusTemporaryRedirect)
		return
	}

	token, err := h.authenticator.Config.Exchange(ctx, r.URL.Query().Get("code"), conf.Oidc.SetValues()...)
	if err != nil {
		h.log.Critical(fmt.Sprintf("no token found: %v", err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		responseError(h.log, w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}

	oidcConfig := &oidc.Config{
		ClientID: conf.Oidc.ClientId,
	}

	_, err = h.authenticator.Provider.Verifier(oidcConfig).Verify(ctx, rawIDToken)

	if err != nil {
		responseError(h.log, w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// var profile map[string]interface{}
	// if err := idToken.Claims(&profile); err != nil {
	// 	responseError(h.log,w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// session.Values["profile"] = profile
	auth.SetTokenSession(session, token)
	err = session.Save(r, w)
	if err != nil {
		responseError(h.log, w, err.Error(), http.StatusInternalServerError)
		return
	}
	redirect, ok := session.Values["redirect"].(string)
	if !ok {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (h *handler) Callback(pattern string) {
	h.mux.Handle(pattern, genInstrumentChain("callback", h.callback))
}

func (h *handler) logout(w http.ResponseWriter, r *http.Request) {
	conf := h.conf
	logoutUrl, err := url.Parse(conf.Oidc.Logout)

	if err != nil {
		responseError(h.log, w, err.Error(), http.StatusInternalServerError)
		return
	}
	session, err := app.Store.Store(conf.ServerName).Get(r, conf.CookieName)
	if err != nil {
		responseError(h.log, w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(r, w)
	app.Store.Store(conf.ServerName).Delete(session)
	http.Redirect(w, r, logoutUrl.String(), http.StatusTemporaryRedirect)
}

func (h *handler) Logout(pattern string) {
	h.mux.Handle(pattern, genInstrumentChain("logout", h.logout))
}

type proxyContextKey struct{}

var proxyKey proxyContextKey

type proxyValue struct {
	registry         *Registry
	host             string
	isProxySslVerify bool
	tokenType        string
	tokenKey         string
}

func fromProxyContext(ctx context.Context) proxyValue {
	return ctx.Value(proxyKey).(proxyValue)
}

func (h *handler) proxy(w http.ResponseWriter, r *http.Request) {
	value := fromProxyContext(r.Context())
	registry := value.registry
	host := value.host
	typ := value.tokenType
	tokenKey := value.tokenKey
	isProxySslVerify := value.isProxySslVerify
	ctx := context.Background()
	conf := h.conf
	log := h.log
	session, err := app.Store.Store(conf.ServerName).Get(r, conf.CookieName)
	if err != nil {
		responseError(h.log, w, err.Error(), http.StatusInternalServerError)
		return
	}
	var rawToken string
	rawToken, err = GetAuthorizarionToken(ctx, tokenKey, conf.Oidc, session)
	if err != nil {
		if err == unAuthorized {
			if conf.Redirect {
				session.Values["redirect"] = r.RequestURI
				session.Save(r, w)
				http.Redirect(w, r, conf.Login, http.StatusTemporaryRedirect)
			} else {
				UnAuthorizedResponse(w, conf.Login)
			}
		} else {
			responseError(h.log, w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	send := func(rawToken string) *Response {
		director := func(req *http.Request) {
			req.URL.Scheme = registry.Endpoint().Scheme
			req.URL.Host = registry.Endpoint().Host
			req.Host = host
		}
		var rt http.RoundTripper = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: isProxySslVerify,
			},
			Dial: func(network, addr string) (net.Conn, error) {
				d := &net.Dialer{}
				for i := 0; i < len(registry.Endpoints); i++ {
					e := registry.Endpoint()
					conn, err := d.Dial(network, e.URL.Host)
					if err != nil {
						continue
					}
					return conn, err
				}

				return nil, errors.ErrNoEndpointsAvailable
			},
		}
		rt = NewDumpTransport(r.Context(), rt)
		rt = NewAuthorizationTransport(typ, rawToken, rt)
		rt = NewIsNotFoundPageTransport(rt)
		reverse := &httputil.ReverseProxy{
			Director:      director,
			ErrorHandler:  errorResponse(log),
			Transport:     rt,
			FlushInterval: -1,
		}
		w, resp := wrapResponseWriter(w)
		reverse.ServeHTTP(w, r)
		return resp
	}
	resp := send(rawToken)
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		rawToken, isSave, err := Token(ctx, h.authenticator, tokenKey, conf.Oidc, session)
		if err != nil {
			if err == unAuthorized {
				if conf.Redirect {
					session.Values["redirect"] = r.RequestURI
					session.Save(r, w)
					http.Redirect(w, r, conf.Login, http.StatusTemporaryRedirect)
				} else {
					UnAuthorizedResponse(w, conf.Login)
				}
			} else {
				responseError(h.log, w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if isSave {
			session.Save(r, w)
		}
		send(rawToken)
	}
}

func (h *handler) Proxy(pattern string, registry *Registry, host, typ, tokenKey string, isProxySslVerify bool) {
	h.mux.Handle(pattern, genInstrumentChain(pattern, func(w http.ResponseWriter, r *http.Request) {
		value := proxyValue{
			registry:         registry,
			host:             host,
			tokenType:        typ,
			tokenKey:         tokenKey,
			isProxySslVerify: isProxySslVerify,
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, proxyKey, value)
		*r = *r.WithContext(ctx)
		h.proxy(w, r)
	}))
}

type Handler interface {
	http.Handler
	Login(pattern string)
	Callback(pattern string)
	Logout(pattern string)
	Proxy(pattern string, registry *Registry, host, typ, tokenKey string, isProxySslVerify bool)
}

func new(conf config.Servers) (Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(rw http.ResponseWriter, r *http.Request) {})
	authenticator, err := auth.NewAuthenticator(context.Background(), conf.Oidc)
	return &handler{
		conf:          conf,
		mux:           mux,
		log:           conf.Logging.GetLogger(),
		authenticator: authenticator,
	}, err
}
