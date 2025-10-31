package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Developer-s-Foundry/df-2.0-aima-auth-service/database/postgres"
	"github.com/Developer-s-Foundry/df-2.0-aima-auth-service/database/rabbitmq"
	"github.com/julienschmidt/httprouter"
)

var ErrUsernameTaken = errors.New("username already exists")

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var user struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := readFromJson(r, &user); err != nil {
		respErr := map[string]string{
			"error":  err.Error(),
			"status": http.StatusText(http.StatusBadRequest),
		}
		writeToJson(w, respErr, http.StatusBadRequest)
		return
	}
	existingUser, err := h.DB.GetUser(r.Context(), user.Email)
	if err != nil && !errors.Is(err, postgres.ErrInvalidUser) {
		log.Printf("unable to get user from db: %v", err)
		respErr := map[string]string{
			"error":  "internal server error",
			"status": http.StatusText(http.StatusInternalServerError),
		}
		writeToJson(w, respErr, http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		respErr := map[string]string{
			"error":  "user already exists",
			"status": http.StatusText(http.StatusConflict),
		}
		writeToJson(w, respErr, http.StatusConflict)
		return
	}

	hashedPassword, _ := hashPassword(user.Password)
	usr := postgres.User{
		UserID:         generateUuid(),
		Email:          user.Email,
		HashedPassword: hashedPassword,
	}

	if err := h.DB.InsertUser(usr); err != nil {
		log.Printf("failed to create user %s: %v", usr.Email, err)
		respErr := map[string]string{
			"error":  "internal server error",
			"status": http.StatusText(http.StatusInternalServerError),
		}
		writeToJson(w, respErr, http.StatusInternalServerError)
		return
	}

	token, err := generateJWToken(usr.UserID)

	if err != nil {
		log.Printf("error generating jwt token %v", err)
		respErr := map[string]string{
			"error":  "internal server error",
			"status": http.StatusText(http.StatusInternalServerError),
		}
		writeToJson(w, respErr, http.StatusInternalServerError)
		return
	}

	userData := map[string]interface{}{
		"data": map[string]string{
			"type":      rabbitmq.NotifyUserSuccessfulSignUp,
			"email":     usr.Email,
			"id":        usr.UserID,
			"timestamp": time.Now().String(),
		},
		"queue_name":    rabbitmq.NotificationQueue,
		"exchange_name": rabbitmq.NotificationExchange,
	}

	// publish to user notification
	go h.RabbMQ.PublishNotification(userData)

	userData = map[string]interface{}{
		"data": map[string]string{
			"type":      rabbitmq.NotifyUserSuccessfulSignUp,
			"email":     usr.Email,
			"id":        usr.UserID,
			"timestamp": time.Now().String(),
		},
		"queue_name":    rabbitmq.UserQueue,
		"exchange_name": rabbitmq.UserExchange,
	}

	//publish to user management
	go h.RabbMQ.PublishUserManagement(userData)

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
		if errors.Is(err, postgres.ErrInvalidUser) {
			respErr := map[string]string{
				"error":  "email or password does not exists",
				"status": http.StatusText(http.StatusBadRequest),
			}
			writeToJson(w, respErr, http.StatusBadRequest)
		} else {
			respErr := map[string]string{
				"error":  "an internal server error occured",
				"status": http.StatusText(http.StatusInternalServerError),
			}
			writeToJson(w, respErr, http.StatusInternalServerError)
		}
		return
	}

	if !checkPasswordHash(authUser.Password, existingUser.HashedPassword) {
		respErr := map[string]string{
			"error":  "invalid login credentials",
			"status": http.StatusText(http.StatusUnauthorized),
		}
		writeToJson(w, respErr, http.StatusUnauthorized)
		return
	}

	sessionToken, err := generateJWToken(existingUser.UserID)

	if err != nil {
		log.Printf("Token generation error for user %s: %v", authUser.Email, err)
		respErr := map[string]string{
			"error":  "internal server error",
			"status": http.StatusText(http.StatusInternalServerError),
		}
		writeToJson(w, respErr, http.StatusInternalServerError)
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
