package config

import "golang.org/x/oauth2"

// Audience
type Audience string

func (a Audience) String() string {
	return string(a)
}

type Audiences []Audience

func (a Audiences) SetValue() []oauth2.AuthCodeOption {
	var audiences []oauth2.AuthCodeOption
	for _, audience := range a {
		audiences = append(audiences, oauth2.SetAuthURLParam("audience", audience.String()))
	}
	return audiences
}
