package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	// we are to generate a uuid to store as part of the user
	// check if user doesn't already exist
	if _, ok := users[username]; ok {
		er := http.StatusConflict
		http.Error(w, "User already exists", er)
		return
	}

	// commiting to database after checking if user doesn't already exist
	// if err := post.Insert(user); err != nil {
	// 	return fmt.Errorf("failed to insert task %s: %w", task.Name, err)
	// }
	hashedPassword, _ := hashPassword(password)
	users[username] = Login{
		HashPassword: hashedPassword,
	}

	fmt.Fprintln(w, "User registered successfully!")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
	// retrieve userid from database pass as string to JWT func
	sessionToken := generateJWToken(32)
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

func (h *AuthHandler) Protected(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
