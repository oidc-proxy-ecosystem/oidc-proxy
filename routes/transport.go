package routes

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/internal"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
)

// var HTTPClient internal.ContextKey
var client *http.Client = http.DefaultClient

func NewContextClient(ctx context.Context, client *http.Client) context.Context {
	return internal.NewContextClient(ctx, client)
}

func FromContextClient(ctx context.Context) *http.Client {
	return internal.FromContextClient(ctx)
}

func NewDumpTransport(ctx context.Context, transport http.RoundTripper) http.RoundTripper {
	if transport == nil {
		transport = FromContextClient(ctx).Transport
	}
	return &DumpTransport{
		Transport: transport,
		log:       logger.FromContext(ctx),
	}
}

func NewClient(ctx context.Context) *http.Client {
	c := *client
	c.Transport = NewDumpTransport(ctx, c.Transport)
	return &c
}

type RoundTrip func(req *http.Request) (*http.Response, error)

type DumpTransport struct {
	log       logger.ILogger
	Transport http.RoundTripper
}

func (t *DumpTransport) transport() http.RoundTripper {
	if t.Transport == nil {
		return http.DefaultTransport
	}
	return t.Transport
}

func (t *DumpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.log.Info(fmt.Sprintf("Connected to %v", req.URL))
	dump := func(b []byte) {
		dumps := strings.Split(string(b), "\n")
		for _, dump := range dumps {
			t.log.Debug(dump)
		}
	}
	// リクエストの送信内容を表示
	b, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	dump(b)
	// 実際のリクエストを送信
	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return nil, err
	}
	// レスポンス内容を表示
	b, err = httputil.DumpResponse(resp, true)
	dump(b)

	return resp, err
}

type AuthorizationTransaport struct {
	Transport http.RoundTripper
	typ       string
	token     string
}

func NewAuthorizationTransport(tokenType, token string, transport http.RoundTripper) http.RoundTripper {
	return &AuthorizationTransaport{
		Transport: transport,
		typ:       tokenType,
		token:     token,
	}
}

func (a *AuthorizationTransaport) transport() http.RoundTripper {
	if a.Transport == nil {
		return http.DefaultTransport
	}
	return a.Transport
}

func (a *AuthorizationTransaport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", a.typ+" "+a.token)
	resp, err := a.transport().RoundTrip(r)
	return resp, err
}

type ErrNotFoundPage struct {
	err error
}

func (this *ErrNotFoundPage) Error() string {
	if this.err != nil {
		return this.err.Error()
	}
	return "Not found page"
}

type isNotFoundPageTransport struct {
	transport http.RoundTripper
}

func NewIsNotFoundPageTransport(transport http.RoundTripper) http.RoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &isNotFoundPageTransport{
		transport: transport,
	}
}

func (this *isNotFoundPageTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := this.transport.RoundTrip(r)
	if resp != nil {
		if resp.StatusCode == http.StatusNotFound {
			return resp, &ErrNotFoundPage{
				err: err,
			}
		}
	}
	return resp, err
}
