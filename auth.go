package twt

import (
	"crypto/subtle"
	"net/http"
)

func BasicAuth(validUsername string, validPassword string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			username, password, _ := req.BasicAuth()
			if verify(validUsername, username, validPassword, password) {
				next(res, req)
			} else {
				http.Error(res, "Unauthorized", http.StatusUnauthorized)
			}
		}
	}
}

func NoAuth() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return next
	}
}

func verify(validUsername, username, validPassword, password string) bool {
	usernameMatches := subtle.ConstantTimeCompare([]byte(validUsername), []byte(username)) == 1
	passwordMatches := subtle.ConstantTimeCompare([]byte(validPassword), []byte(password)) == 1
	return usernameMatches && passwordMatches
}
