# oidc-proxy

## 概要

OpenID Connectへ認証を行い、プロキシ先へ`Authorization`ヘッダーに設定された認証方式を付与してプロキシ先に転送する。

### 認証方式

OAuth2: Authorization Code Flowに基づくOpenID Connect

## Contents

- [oidc-proxy](#oidc-proxy)
  - [概要](#概要)
    - [認証方式](#認証方式)
  - [Contents](#contents)
  - [application config file](#application-config-file)
    - [TopLevel](#toplevel)
    - [Logging](#logging)
    - [servers](#servers)
    - [oidc](#oidc)
    - [location](#location)
    - [urls](#urls)
    - [session](#session)
    - [example](#example)

## application config file

サポートしているファイル拡張子

- .yaml
- .yml
- .toml
- .json

設定ファイル内に環境変数を設定することも可能です。

設定ファイルについてはファイル監視を行い、変更が入り次第設定を読込

ルーティング設定等々の再設定が行われます。

ssl証明書ファイルに関しても、ファイル監視を行っているため

再起動なしで再設定が可能です。

### TopLevel

| キー                | タイプ | 内容                         | required |
| :------------------ | :----: | :--------------------------- | :------: |
| logging             | object | [Logging](#logging)          |   true   |
| servers             | array  | [Servers](#servers)          |   true   |
| port                | number | プロキシサーバーのポート番号 |   true   |
| ssl_certificate     | string | .crtファイル                 |  false   |
| ssl_certificate_key | string | .keyファイル                 |  false   |

### Logging

| キー       | タイプ | 内容                                                                               | required | default  |
| :--------- | :----: | :--------------------------------------------------------------------------------- | :------: | :------- |
| level      | string | ログの出力レベルが設定出来ます。(debug, info, warn, error, criticalのどれかを設定) |  false   | default  |
| filename   | string | 出力先ファイル名を設定(絶対パス)                                                   |  false   |          |
| logformat  | string | ログの出力フォーマットの設定(short, standard, long)                                |  false   | standard |
| timeformat | string | ログ出力時の時間フォーマットの設定(date, datetime, millisec)                       |  false   | datetime |

### servers

| キー        | タイプ  | 内容                                                            | required |
| :---------- | :-----: | :-------------------------------------------------------------- | :------: |
| oidc        | object  | [OIDC](#oidc)                                                   |   true   |
| locations   |  array  | [Location](#location)                                           |   true   |
| logging     | object  | [Logging](#logging)                                             |   true   |
| session     | object  | [Session](#session)                                             |   true   |
| cookie_name | string  | cookieセッション名                                              |   true   |
| server_name | string  | バーチャルホスト名                                              |   true   |
| port        |   int   |                                                                 |   true   |
| login       | string  | プロキシサーバー上のログインURLを設定                           |   true   |
| callback    | string  | プロキシサーバー上のコールバックURL                             |   true   |
| logout      | string  | プロキシサーバー上のログアウトURL                               |   true   |
| redirect    | boolean | セッション情報が消失した際にログインURLへリダレクトするかどうか |  false   |

### oidc

| キー          | タイプ | 内容                            | required |
| :------------ | :----: | :------------------------------ | :------: |
| scopes        | array  | oidcスコープ                    |   true   |
| provider      | string | プロバイダURL                   |   true   |
| client_id     | string | IDPクライアントキー             |   true   |
| client_secret | string | IDPクライアントシークレットキー |   true   |
| redirect_url  | string | リダイレクトURL                 |   true   |
| logout        | string | IDPのログアウト先URL            |   true   |

### location

| キー       | タイプ | 内容         | required |
| :--------- | :----: | :----------- | :------: |
| proxy_pass | string | 転送先URL    |   true   |
| urls       | array  | [URL](#urls) |   true   |

### urls

| キー  | タイプ | 内容                                                       | required |
| :---- | :----: | :--------------------------------------------------------- | :------: |
| path  | string | 転送先URLパス                                              |   true   |
| token | string | 転送先パスへ転送するトークンを設定(id_token, access_token) |   true   |

### session

| キー       | タイプ | 内容                                         |  required  |
| :--------- | :----: | :------------------------------------------- | :--------: |
| name       | string | 使用するキャッシュプラグイン名               |    true    |
| codecs     | array  | Cookieセッションを暗号化するためのキー文字列 |    true    |
| cache_time | number | キャッシュを保持しておく時間                 |    true    |
| endpoints  | array  | キャッシュサーバーへの接続先エンドポイント   | true(etcd) |
| username   | string | キャッシュサーバーへの接続ユーザー名         |   false    |
| password   | string | キャッシュサーバーへの接続パスワード         |   false    |

### example

```yaml
port: 8080
ssl_certificate: ssl/sever.crt
ssl_certificate_key: ssl/sever.key
logging:
    level: debug or info or warn or warning(warn) or error or err(error) or critical or dev(debug) or prod(info)
    filename: ""
    prefix: ""
servers:
    - oidc:
        scopes:
            - email
            - openid
            - offline_access
            - profile
        provider: https://keycloak/
        client_id: xxx
        client_secret: xxx
        redirect_url: http://localhost:8080/oauth2/callback
        logout: https://keycloak/logout?returnTo=http://localhost:8080/oauth2/login
      login: /oauth2/login
      callback: /oauth2/callback
      logout: /oauth2/logout
      locations:
        - proxy_pass: http://localhost
          urls:
            - path: /
              token: id_token
      logging:
        level: debug or info or warn or warning(warn) or error or err(error) or critical or dev(debug) or prod(info)
        filename: ""
        prefix: ""
      session_name: ""
      server_name: virtual sever name
      cache:
        name: memory
        codecs: []
        endpoints: []
        cache_time: 30
        username: ""
        password: ""
```
