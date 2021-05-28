package validation

import "github.com/oidc-proxy-ecosystem/oidc-proxy/translate"

var langs = []struct {
	lang translate.LocalType
	msg  map[string]string
}{
	{
		lang: translate.Japan,
		msg: map[string]string{
			required: "{0}は必須フィールドです。",
		},
	},
	{
		lang: translate.English,
		msg: map[string]string{
			required: "{0} is required field.",
		},
	},
}

func init() {
	for _, lang := range langs {
		translate.RegisterLanguageType(lang.lang, lang.msg)
	}
}
