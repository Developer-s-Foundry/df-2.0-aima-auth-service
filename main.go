package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
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
	_ = godotenv.Load()

	portString := os.Getenv("PORT")
	rConnStr := os.Getenv("RABBITMQ_URL")
	if portString == "" {
		log.Fatal("PORT not found in environment variables!")
	}

	portInt, err := strconv.Atoi(portString)
	if err != nil {
		log.Fatal("Invalid PORT parameter")
	}

	url, user := os.Getenv("DB_URL"), os.Getenv("DB_USER")
	host := os.Getenv("DB_HOST")
	password, port := os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT")
	dbName, dbSSL := os.Getenv("DB_NAME"), os.Getenv("DB_SSL")

	post, err := postgres.ConnectPostgres(url, password, port, host, dbName, user, dbSSL)
	if err != nil {
		panic(err)
	}

	rabbit := rabbitmq.NewRabbitMQ(rConnStr)

	var wg sync.WaitGroup
	// ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := rabbit.Connect(); err != nil {
			log.Printf("[RabbitMQ] Connection error: %v", err)
		}
	}()

	<-rabbit.NotifyReady()
	if err := rabbit.DeclareExchangesAndQueues(); err != nil {
		log.Fatal(err)
	}

	// consumeEmail := rabbitmq.NewConsumer(rabbit, "notification_email_queue", "email-worker", handleEmailDeliveryAck)

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	consumeEmail.Start(ctx)
	// }()

	auth := &AuthHandler{DB: post, RabbMQ: rabbit}
	router := httprouter.New()
	router.POST("/register", VerifyGatewayRequest(auth.Register))
	router.POST("/login", VerifyGatewayRequest(auth.Login))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", portInt),
		ReadTimeout:  time.Minute * 30,
		WriteTimeout: time.Minute * 30,
		Handler:      router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Server is running on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	<-stop
	log.Println("[Main] Shutdown signal received")

	// cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Main] Server shutdown error: %v", err)
	} else {
		log.Println("[Main] HTTP server shut down gracefully")
	}

	rabbit.Close()

	wg.Wait()
	log.Println("[Main] All goroutines exited cleanly")
}
