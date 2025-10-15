package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var ErrAuth = errors.New("Unauthorized")

func AuthorizeRequest(r *http.Request) (*CustomClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrAuth
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, ErrAuth
	}
	sessionToken := parts[1]

	claims := &CustomClaims{}

	// validate the JWT here
	token, err := jwt.ParseWithClaims(sessionToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtRSAPublicKey, nil
	})

	if err != nil {
		log.Printf("JWT validation failed: %v", err)
		return nil, ErrAuth
	}

	if !token.Valid {
		return nil, ErrAuth
	}
	return claims, nil
}
