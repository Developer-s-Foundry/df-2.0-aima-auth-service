package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Developer-s-Foundry/df-2.0-aima-auth-service/database/postgres"
	"github.com/julienschmidt/httprouter"
)

var ErrUsernameTaken = errors.New("username already exists")

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid method", er)
		return

	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	if len(email) < 8 || len(password) < 8 {
		er := http.StatusNotAcceptable
		http.Error(w, "Invalid email/password", er)
		return
	}

	// check if user doesn't already exist
	existingUser, err := h.DB.GetUser(r.Context(), email)
	if existingUser != nil {
		writeToJson(w, "User already exists", http.StatusConflict)
		return
	} else {
		log.Printf("unable to get user from db: %v", err)
	}

	// commit to database after checking if user doesn't already exist
	hashedPassword, _ := hashPassword(password)
	user := postgres.User{
		UserID:         generateUuid(),
		Email:          email,
		HashedPassword: hashedPassword,
		CreatedAt:      time.Now(),
	}

	if err := h.DB.InsertUser(user); err != nil {
		fmt.Printf("failed to create user %s: %v", email, err)
		writeToJson(w, "Failed to create user", http.StatusInternalServerError)
	}

	response := struct {
		StatusCode int    `json:"status_code"`
		UserId     string `json:"userId"`
		Message    string `json:"message"`
	}{
		StatusCode: http.StatusCreated,
		UserId:     user.UserID,
		Message:    "User created successfully",
	}
	writeToJson(w, response, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid Request Method", er)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	existingUser, err := h.DB.GetUser(r.Context(), email)

	if err != nil {
		log.Printf("DB error for user %s: %v", email, err)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "An internal server error occurred", http.StatusInternalServerError)
		return
	}

	if !checkPasswordHash(password, existingUser.HashedPassword) {
		http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	jwtData := map[string]interface{}{
		"email":   email,
		"user_id": existingUser.UserID,
		"role_id": existingUser.RoleId,
	}
	sessionToken, err := generateJWToken(jwtData)

	if err != nil {
		log.Printf("Token generation error for user %s: %v", email, err)
		http.Error(w, "Cannot generate token", http.StatusInternalServerError)
		return
	}

	response := struct {
		Token      string `json:"session_token"`
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}{
		Token:      sessionToken,
		StatusCode: http.StatusOK,
		Message:    "Login successful",
	}

	writeToJson(w, response, http.StatusOK)
	fmt.Fprintln(w, "Login successful!")

}

func (h *AuthHandler) UpdateUsername(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	claims, err := AuthorizeRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	newUsername := r.FormValue("new_username")

	if strings.TrimSpace(newUsername) == "" || len(newUsername) < 4 {
		http.Error(w, "New username is required and must be at least 4 characters long", http.StatusBadRequest)
		return
	}

	user_id := claims.UserId
	email := claims.Email

	if err := h.DB.UpdateUsername(r.Context(), user_id, newUsername); err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			http.Error(w, "This username is already taken.", http.StatusConflict)
			return
		}

		log.Printf("DB error updating username for user %s: %v", email, err)
		http.Error(w, "Failed to update username due to a server error.", http.StatusInternalServerError)
		return
	}

	// fix this: claims should have userId and a roleId
	newJwtData := map[string]interface{}{
		"email":   newUsername,
		"user_id": claims.UserId,
		"role_id": claims.RoleId,
	}

	newToken, err := generateJWToken(newJwtData)
	if err != nil {
		log.Printf("Token generation error after username update: %v", err)
		http.Error(w, "Username updated, but failed to generate new token.", http.StatusInternalServerError)
		return
	}

	response := struct {
		Token      string `json:"data"`
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}{
		Token:      newToken,
		StatusCode: http.StatusOK,
		Message:    "Username updated successfully. Please use the new token for subsequent requests.",
	}

	writeToJson(w, response, http.StatusOK)
}

func (h *AuthHandler) Protected(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid Request Method", er)
		return
	}

	claims, err := AuthorizeRequest(r)

	if err != nil {
		er := http.StatusUnauthorized
		http.Error(w, "Unauthorised", er)
		return
	}

	fmt.Fprintf(w, "Login successful! Welcome user %s ", claims.Email)
}
