package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

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

	token, err := generateJWToken(usr.UserID)

	if err != nil {
		log.Printf("error generating jwt token %v", err)
		writeToJson(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		UserId     string `json:"userId"`
		Email      string `json:"email"`
		Message    string `json:"message"`
		Token      string `json:"token"`
		StatusCode int    `json:"status_code"`
	}{

		StatusCode: http.StatusCreated,
		UserId:     usr.UserID,
		Email:      usr.Email,
		Token:      token,
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

	sessionToken, err := generateJWToken(existingUser.UserID)

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
