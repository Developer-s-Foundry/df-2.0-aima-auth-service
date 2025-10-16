package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Developer-s-Foundry/df-2.0-aima-auth-service/database/postgres"
	"github.com/Developer-s-Foundry/df-2.0-aima-auth-service/database/rabbitmq"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

type AuthHandler struct {
	DB     *postgres.PostgresConn
	RabbMQ *rabbitmq.RabbitMQ
}

func main() {
	godotenv.Load()

	portString := os.Getenv("PORT")
	rConnStr := os.Getenv("RABBITMQ_URL")
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

	rabbit := rabbitmq.NewRabbitMQ(rConnStr)

	go rabbit.Connect()
	<-rabbit.NotifyReady()

	auth := &AuthHandler{DB: post, RabbMQ: rabbit}
	router := httprouter.New()
	router.POST("/auth/register", auth.Register)
	router.POST("/auth/login", auth.Login)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", portInt),
		ReadTimeout:  time.Minute * 30,
		WriteTimeout: time.Minute * 30,
		Handler:      router,
	}
	log.Printf("Server is running on %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
