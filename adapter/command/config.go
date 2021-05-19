package command

import (
	"github.com/oidc-proxy-ecosystem/oidc-proxy/config"
	"github.com/urfave/cli/v2"
)

var AppFileCommand = &cli.Command{
	Name:    "config",
	Aliases: []string{"c"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Value:   "application.yml",
		},
	},
	Action: func(c *cli.Context) error {
		filename := c.String("output")
		example := config.Config{
			Logging: config.Logging{
				Level: "debug or info or warn or warning(warn) or error or err(error) or critical or dev(debug) or prod(info)",
			},
			Port:              8080,
			SslCertificate:    "ssl/sever.crt",
			SslCertificateKey: "ssl/sever.key",
			Servers: []*config.Servers{
				{
					ServerName: "virtual sever name",
					Login:      "/oauth2/login",
					Callback:   "/oauth2/callback",
					Logout:     "/oauth2/logout",
					Logging: config.Logging{
						Level: "debug or info or warn or warning(warn) or error or err(error) or critical or dev(debug) or prod(info)",
					},
					Oidc: config.Oidc{
						Scopes:       []string{"email", "openid", "offline_access", "profile"},
						Provider:     "https://keycloak/",
						ClientId:     "xxx",
						ClientSecret: "xxx",
						Logout:       "https://keycloak/logout?returnTo=http://localhost:8080/oauth2/login",
						RedirectUrl:  "http://localhost:8080/oauth2/callback",
					},
					Locations: []config.Locations{
						{
							ProxyPass: "http://localhost",
							Urls: []config.Urls{
								{
									Path:  "/",
									Token: "id_token",
								},
							},
						},
					},
					Session: config.Session{
						Name:   "memory",
						Codecs: []string{},
						Args: map[string]interface{}{
							"endpoints": []string{""},
							"username":  "",
							"password":  "",
							"ttl":       30,
						},
					},
				},
			},
		}
		return example.Output(filename)
	},
}
