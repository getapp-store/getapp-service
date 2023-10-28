package utils

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

const (
	AppKey  = "X-App-Key"
	UserKey = "X-User-Key"
)

func TokenFromHeader(r *http.Request, key string) string {
	// Get token from authorization header.
	bearer := r.Header.Get(key)
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}

func TokenFromParams(r *http.Request, key string) string {
	// Get token from authorization header.
	var token = r.URL.Query().Get(strings.ToLower(key))
	if token == "" {
		token = chi.URLParam(r, key)
	}
	return token
}
