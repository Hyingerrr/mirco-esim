package rest

import "net/http"

type (
	Middleware func(next http.HandlerFunc) http.HandlerFunc

	Route struct {
		Method  string
		Path    string
		Handler http.HandlerFunc
	}
)
