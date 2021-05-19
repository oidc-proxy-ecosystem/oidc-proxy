package routes

import (
	"context"
	"errors"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/sessions"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/auth"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/config"
	"golang.org/x/oauth2"
)

var (
	unAuthorized = errors.New("authorization error")
	noTokenKey   = errors.New("no token key")
)

func Token(ctx context.Context, tokenKey string, oidcConf config.Oidc, session *sessions.Session) (string, bool, error) {
	var rawToken string
	var isSave bool = false
	var resultErr error
	rawIdToken, ok := session.Values["id_token"].(string)
	if !ok {
		return rawToken, isSave, unAuthorized
	}
	if rawIdToken != "" {
		authenticator, err := auth.NewAuthenticator(ctx, oidcConf)
		if err != nil {
			return "", isSave, err
		}
		oidcConfig := &oidc.Config{
			ClientID: oidcConf.ClientId,
		}
		// IDトークンの検証
		_, err = authenticator.Provider.Verifier(oidcConfig).Verify(ctx, rawIdToken)
		if err != nil {
			// トークンの更新
			refreshToken := session.Values["refresh_token"].(string)
			ts := authenticator.Config.TokenSource(ctx, &oauth2.Token{
				RefreshToken: refreshToken,
			})
			if token, err := ts.Token(); err != nil {
				return "", false, err
			} else {
				auth.SetTokenSession(session, token)
				isSave = true
			}
			// session.Save(r, w)
		}
		// プロキシ先へ転送するトークンを取得
		rawToken, ok = session.Values[tokenKey].(string)
		if !ok {
			resultErr = noTokenKey
			rawToken = rawIdToken
		}
	} else {
		return "", false, unAuthorized
	}
	return rawToken, isSave, resultErr
}
