package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	auth "github.com/stedankyi/df-2.0-aima-auth-service/handlers"
)

type Login struct {
	HashPassword string
	SessionToken string
}

var users = map[string]Login{}

func main() {
	fmt.Println("Hello, Welcome to the Church Management System!")

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

	fmt.Println("Port:", portString)

	// endpoints and handlers
	router := httprouter.New()
	router.POST("/api/v1/login", auth.Login)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", portInt),
		ReadTimeout:  time.Minute * 30,
		WriteTimeout: time.Minute * 30,
		Handler:      router,
	}
	log.Printf("Server is running on %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
