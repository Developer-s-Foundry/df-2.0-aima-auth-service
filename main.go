package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Developer-s-Foundry/df-2.0-aima-auth-service/database/postgres"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

type AuthHandler struct {
	DB *postgres.PostgresConn
}

// roles
type RoleId string

const (
	RoleAnalyst       RoleId = "Analyst"
	RoleManager       RoleId = "Manager"
	RoleDeveloper     RoleId = "Developer"
	RoleAdministrator RoleId = "Administrator"
)

func main() {
	fmt.Println("Hello, Welcome to AIMA AuthService!")

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment file!")
	}

	portInt, err := strconv.Atoi(portString)
	if err != nil {
		log.Fatal("Invalid port parameter passed")
	}

	// database setup
	url, user := os.Getenv("DB_URL"), os.Getenv("DB_USER")
	host := os.Getenv("DB_HOST")
	password, port := os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT")
	db_name, db_ssl := os.Getenv("DB_NAME"), os.Getenv("DB_SSL")

	fmt.Println("Port:", portString)

	post, err := postgres.ConnectPostgres(url, password, port, host, db_name, user, db_ssl)
	if err != nil {
		panic(err)
	}

	// endpoints and handlers
	auth := &AuthHandler{DB: post}
	router := httprouter.New()
	router.POST("/api/v1/register", auth.Register)
	router.POST("/api/v1/login", auth.Login)
	router.POST("/api/v1/protected", auth.Protected)
	router.POST("/api/v1/update", auth.UpdateUsername)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", portInt),
		ReadTimeout:  time.Minute * 30,
		WriteTimeout: time.Minute * 30,
		Handler:      router,
	}
	log.Printf("Server is running on %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
