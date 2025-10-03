package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid method", er)
		return

	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if len(username) < 8 || len(password) < 8 {
		er := http.StatusNotAcceptable
		http.Error(w, "Invalid username/password", er)
		return
	}

	if _, ok := users[username]; ok {
		er := http.StatusConflict
		http.Error(w, "User already exists", er)
		return
	}
	hashedPassword, _ := hashPassword(password)
	users[username] = Login{
		HashPassword: hashedPassword,
	}

	fmt.Fprintln(w, "User registered successfully!")
}

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid Request Method", er)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, ok := users[username] // user in-memory database for users

	if !ok || !checkPasswordHash(password, user.HashPassword) {
		er := http.StatusUnauthorized
		http.Error(w, "Invalid username or password", er)
		return
	}

	// generate session token
	sessionToken := generateSessionToken(32)
	user.SessionToken = sessionToken
	users[username] = user

	response := struct {
		Token      string `json:"data"`
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}{
		Token:      sessionToken,
		StatusCode: http.StatusOK,
		Message:    "task updated successfully",
	}

	writeToJson(w, response, http.StatusOK)
	fmt.Fprintln(w, "Login successful!")
}

func protected(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid Request Method", er)
		return
	}

	if err := Authorize(r); err != nil {
		er := http.StatusUnauthorized
		http.Error(w, "Unauthorised", er)
		return
	}

	username := r.FormValue("username")

	fmt.Fprintf(w, "Login successful! Welcome %s", username)
}
