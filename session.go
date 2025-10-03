package main

import (
	"errors"
	"net/http"
)

var ErrAuth = errors.New("Unauthorized")

func Authorize(r *http.Request) error {
	username := r.FormValue("username")
	user, ok := users[username]
	if !ok {
		return ErrAuth
	}

	// session token
	sessionToken := r.Header.Get("Auth-Token")
	if sessionToken != user.SessionToken || sessionToken == "" {
		return ErrAuth
	}
	return nil
}
