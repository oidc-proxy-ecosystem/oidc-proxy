package translate

type LocalType string

const (
	Japan   = LocalType("ja")
	English = LocalType("en")
)

var translate = make(map[LocalType]map[string]string)

func RegisterLanguageType(lang LocalType, msg map[string]string) {
	translate[lang] = msg
}

func New(lang LocalType) map[string]string {
	if msg, ok := translate[lang]; ok {
		return msg
	}
	return map[string]string{}
}
