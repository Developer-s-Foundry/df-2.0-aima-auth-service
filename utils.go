package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CustomClaims struct {
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

type JWtClaim struct {
	UserID string
}

func generateJWToken(data JWtClaim) (string, error) {
	user_id := data.UserID
	jwtIssuer := os.Getenv("JWT_ISSUER")

	claims := CustomClaims{
		UserId: user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    jwtIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(jwtRSAPrivateKey)

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

func readFromJson(r *http.Request, receiver interface{}) error {
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(receiver); err != nil {
		return fmt.Errorf("failed to read from user parsed response %w", err)
	}
	return nil
}

func generateUuid() string {
	return uuid.NewString()
}

func initRSAKeys(privateKeyPath, publicKeyPath string) error {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("could not read private key file: %w", err)
	}

	// parse the private key
	jwtRSAPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return fmt.Errorf("could not parse private key: %w", err)
	}

	// paes public key
	publicKeyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("could not read public key file: %w", err)
	}

	jwtRSAPublicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		return fmt.Errorf("could not parse public key: %w", err)
	}

	log.Println("RSA Private Key loaded successfully.")
	return nil
}
