package internal

import (
	"context"
	"net/http"
)

var HTTPClient ContextKey

type ContextKey struct{}

func NewContextClient(ctx context.Context, client *http.Client) context.Context {
	return context.WithValue(ctx, HTTPClient, client)
}

func FromContextClient(ctx context.Context) *http.Client {
	if val, ok := ctx.Value(HTTPClient).(*http.Client); ok {
		return val
	}
	return http.DefaultClient
}
