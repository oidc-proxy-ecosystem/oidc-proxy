package routes

import (
	"net/http"
)

func init() {
	html := `<html>
	<head>
		<title>{{.Title}}</title>
	</head>
	<body>
		<p>Page not found</p>
		<p>{{.Err}}<p>
	</body>
</html>`
	MustRegistryPages(http.StatusNotFound, html)
}
