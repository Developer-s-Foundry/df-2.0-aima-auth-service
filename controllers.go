package main

import (
	"fmt"
	"net/http"
)

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid Request Method", er)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, ok := users[username] // add in memory database for users

	if !ok || !checkPasswordHash(password, user.HashPassword) {
		er := http.StatusUnauthorized
		http.Error(w, "Invalid username or password", er)
		return
	}

	// // generate session token

	// sessionToken := generateSessionToken(32)

	// http.SetCookie(w, &http.Cookie)

	fmt.Fprintln(w, "Login successful!")

}
func getUser(w http.ResponseWriter, r *http.Request)    {}
func getUsers(w http.ResponseWriter, r *http.Request)   {}
func createUser(w http.ResponseWriter, r *http.Request) {}
func updateUser(w http.ResponseWriter, r *http.Request) {}
func deleteUser(w http.ResponseWriter, r *http.Request) {}
