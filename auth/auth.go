package auth

import (
	"context"
	"net/url"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/config"

	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	Provider *oidc.Provider
	Config   oauth2.Config
	Ctx      context.Context
}

func (a *Authenticator) setValue(url.Values) {

}

func NewAuthenticator(ctx context.Context, oidcConf config.Oidc) (*Authenticator, error) {
	provider, err := oidc.NewProvider(ctx, oidcConf.Provider)
	if err != nil {
		return nil, err
	}
	var o2conf oauth2.Config
	o2conf = oauth2.Config{
		ClientID:     oidcConf.ClientId,
		ClientSecret: oidcConf.ClientSecret,
		RedirectURL:  oidcConf.RedirectUrl,
		Endpoint:     provider.Endpoint(),
		Scopes:       oidcConf.Scopes,
	}

	return &Authenticator{
		Provider: provider,
		Config:   o2conf,
		Ctx:      ctx,
	}, nil
}

func SetTokenSession(session *sessions.Session, token *oauth2.Token) {
	rawIdToken, _ := token.Extra("id_token").(string)
	session.Values["id_token"] = rawIdToken
	session.Values["access_token"] = token.AccessToken
	session.Values["refresh_token"] = token.RefreshToken
}
