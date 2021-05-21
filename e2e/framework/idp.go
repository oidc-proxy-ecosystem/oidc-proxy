package framework

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	mrand "math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/e2e/utils"
	"golang.org/x/oauth2/jws"
	"gopkg.in/square/go-jose.v2"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type provider struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSEndpoint          string `json:"jwks_uri"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
}

type token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	IdToken      string `json:"id_token"`
}

type IdentityProvider struct {
	*http.Server
	Issuer         string
	PrivateKey     *rsa.PrivateKey
	codes          map[string]struct{}
	idleConnsClose chan struct{}
}

type context struct {
	writer http.ResponseWriter
	req    *http.Request
}

type middleware func(c *context)

func NewIdpServer(idleConnsClose chan struct{}) (*IdentityProvider, error) {
	port, err := utils.FindPort()
	if err != nil {
		return nil, err
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	issuer := fmt.Sprintf("http://127.0.0.1:%d", port)
	idp := &IdentityProvider{
		Issuer:     issuer,
		PrivateKey: privateKey,
		codes:      map[string]struct{}{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(rw http.ResponseWriter, r *http.Request) {
		p := &provider{
			Issuer:                issuer,
			AuthorizationEndpoint: issuer + "/authorize",
			TokenEndpoint:         issuer + "/oauth/token",
			JWKSEndpoint:          issuer + "/.well-known/jwks.json",
		}
		if err := json.NewEncoder(rw).Encode(p); err != nil {
			return
		}
	})
	mux.HandleFunc("/authorize", idp.middleware(idp.handlerAuthorize))
	mux.HandleFunc("/u/login", idp.middleware(idp.handlerLogin))
	mux.HandleFunc("/oauth/token", idp.middleware(idp.handleToken))
	mux.HandleFunc("/.well-known/jwks.json", idp.middleware(idp.handleJWKS))

	idp.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	l, err := net.Listen("tcp", idp.Server.Addr)
	if err != nil {
		return nil, err
	}
	go func() {
		if err := idp.Server.Serve(l); err != http.ErrServerClosed {
			log.Println(err)
		}
		close(idleConnsClose)
	}()
	return idp, nil
}

type AuthResponse struct {
	Query    string
	LoginURL string
}

func (i *IdentityProvider) middleware(handler middleware) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := &context{
			writer: w,
			req:    r,
		}
		handler(c)
	}
}

func (i *IdentityProvider) handlerAuthorize(c *context) {
	http.Redirect(c.writer, c.req, fmt.Sprintf("%s/%s?%s", i.Issuer, "u/login", c.req.URL.Query().Encode()), http.StatusTemporaryRedirect)
}

func (i *IdentityProvider) handlerLogin(c *context) {
	w := c.writer
	r := c.req
	q := r.URL.Query()
	redirectURL, err := url.Parse(q.Get("redirect_uri"))
	if strings.HasPrefix(redirectURL.String(), "http://127.0.0.1:8888/oauth2/callback") {
		log.Panicln(redirectURL.String())
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	code := make([]byte, 16)
	for i := range code {
		code[i] = letters[mrand.Intn(len(letters))]
	}
	i.codes[string(code)] = struct{}{}
	rq := redirectURL.Query()
	rq.Set("state", q.Get("state"))
	rq.Set("code", string(code))
	redirectURL.RawQuery = rq.Encode()
	w.Header().Set("Location", redirectURL.String())
	w.WriteHeader(http.StatusSeeOther)
}

func (i *IdentityProvider) handleToken(c *context) {
	w := c.writer
	r := c.req
	if err := r.ParseForm(); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	gotCode := r.FormValue("code")
	if _, ok := i.codes[gotCode]; !ok {
		log.Print("Unknown code")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cs := &jws.ClaimSet{
		Iss:           i.Issuer,
		Aud:           "oidc-proxy-ecosystem-provider",
		PrivateClaims: map[string]interface{}{"email": "oidc-proxy-ecosystem@n-creativesystem.dev"},
	}
	idToken, err := jws.Encode(&jws.Header{Algorithm: "RS256", KeyID: "idp"}, cs, i.PrivateKey)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	token := &token{
		AccessToken:  "accesstoken",
		RefreshToken: "refreshtoken",
		IdToken:      idToken,
	}
	if err := json.NewEncoder(w).Encode(token); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (i *IdentityProvider) handleJWKS(c *context) {
	w := c.writer
	jwks := &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{Key: i.PrivateKey.Public(), KeyID: "idp"},
		},
	}
	if err := json.NewEncoder(w).Encode(jwks); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
