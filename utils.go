package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CustomClaims struct {
	Email  string `json:"email"`
	UserId string `json:"user_id"`
	RoleId string `json:"role_id"`
	jwt.RegisteredClaims
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateJWToken(data map[string]interface{}) (string, error) {
	email, email_ok := data["email"].(string)
	user_id, user_id_ok := data["user_id"].(string)
	role_id, role_id_ok := data["role_id"].(string)
	if !email_ok {
		return "", errors.New("jwtData must contain a valid 'email' string")
	} else if !user_id_ok {
		return "", errors.New("jwtData must contain a valid 'user_id' string")
	} else if !role_id_ok {
		return "", errors.New("jwtData must contain a valid 'role_id' string")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtIssuer := os.Getenv("JWT_ISSUER")

	claims := CustomClaims{
		Email:  email,
		UserId: user_id,
		RoleId: role_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuer, // needs to be defined in env
			Audience:  jwt.ClaimStrings{email},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret) // needs to be defined in env

	if err != nil {
		log.Printf("Error signing JWT: %v", err)
		return "", errors.New("error generating JWT")
	}
	return tokenString, nil
}

func writeToJson(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}
	return nil
}

func generateUuid() string {
	return uuid.NewString()
}

func validateRole(input string) (RoleId, bool) {

	lowerInput := strings.ToLower(input)

	var roleMap = map[string]RoleId{
		"analyst":       RoleAnalyst,
		"manager":       RoleManager,
		"developer":     RoleDeveloper,
		"administrator": RoleAdministrator,
	}

	roleId, ok := roleMap[lowerInput]
	return roleId, ok
}
