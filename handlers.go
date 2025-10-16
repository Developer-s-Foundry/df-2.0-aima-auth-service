package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Developer-s-Foundry/df-2.0-aima-auth-service/database/postgres"
	"github.com/julienschmidt/httprouter"
)

var ErrUsernameTaken = errors.New("username already exists")

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var user struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := readFromJson(r, &user); err != nil {
		writeToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	existingUser, err := h.DB.GetUser(r.Context(), user.Email)
	if existingUser != nil {
		writeToJson(w, "User already exists", http.StatusConflict)
		return
	} else {
		log.Printf("unable to get user from db: %v", err)
	}

	// commit to database after checking if user doesn't already exist
	hashedPassword, _ := hashPassword(user.Password)
	usr := postgres.User{
		UserID:         generateUuid(),
		Email:          user.Email,
		HashedPassword: hashedPassword,
	}

	if err := h.DB.InsertUser(usr); err != nil {
		fmt.Printf("failed to create user %s: %v", usr.Email, err)
		writeToJson(w, "Failed to create user", http.StatusInternalServerError)
		return
	}


	
	token := generateJWToken()

	response := struct {
		UserId     string `json:"userId"`
		Email      string `json:"email"`
		Message    string `json:"message"`
		StatusCode int    `json:"status_code"`
	}{

		StatusCode: http.StatusCreated,
		UserId:     usr.UserID,
		Email:      usr.Email,
		Message:    "User created successfully",
	}
	writeToJson(w, response, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var authUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := readFromJson(r, &authUser); err != nil {
		writeToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	existingUser, err := h.DB.GetUser(r.Context(), authUser.Email)

	if err != nil {
		log.Printf("DB error for user %s: %v", authUser.Email, err)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "An internal server error occurred", http.StatusInternalServerError)
		return
	}

	if !checkPasswordHash(authUser.Password, existingUser.HashedPassword) {
		http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	jwtData := map[string]interface{}{
		"user_id": existingUser.UserID,
	}
	sessionToken, err := generateJWToken(jwtData)

	if err != nil {
		log.Printf("Token generation error for user %s: %v", authUser.Email, err)
		http.Error(w, "Cannot generate token", http.StatusInternalServerError)
		return
	}

	response := struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		Token      string `json:"session_token"`
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}{
		ID:         existingUser.UserID,
		Email:      existingUser.Email,
		Token:      sessionToken,
		StatusCode: http.StatusOK,
		Message:    "Login successful! Welcome to AIMAS",
	}

	writeToJson(w, response, http.StatusOK)
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

	fmt.Fprintf(w, "Authorisation successful! Welcome user %s ", claims.Email)
}
