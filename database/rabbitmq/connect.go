package rabbitmq

import (
	"log"
	"math"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	URL           string
	conn          *amqp.Connection
	channel       *amqp.Channel
	notifyClose   chan *amqp.Error
	notifyConnErr chan *amqp.Error
	done          chan bool
	ready         chan bool
	mu            sync.Mutex
}

func NewRabbitMQ(url string) *RabbitMQ {
	return &RabbitMQ{
		URL:   url,
		done:  make(chan bool),
		ready: make(chan bool, 1),
	}
}

func (r *RabbitMQ) Connect() error {
	var err error
	for attempt := 0; ; attempt++ {
		r.conn, err = amqp.Dial(r.URL)
		if err == nil {
			r.channel, err = r.conn.Channel()
			if err == nil {
				r.notifyClose = make(chan *amqp.Error)
				r.notifyConnErr = r.conn.NotifyClose(make(chan *amqp.Error))
				log.Println("[RabbitMQ] Connected successfully")

				select {
				case r.ready <- true:
				default:
				}

				go r.handleReconnect()
				return nil
			}
		}

		backoff := exponentialBackoff(attempt)
		log.Printf("[RabbitMQ] Connect failed: %v — retrying in %.1fs", err, backoff.Seconds())
		time.Sleep(backoff)
	}
}

func (r *RabbitMQ) handleReconnect() {
	for {
		select {
		case err := <-r.notifyConnErr:
			if err != nil {
				log.Printf("[RabbitMQ] Connection closed: %v — reconnecting...", err)
				r.reconnectLoop()
			}
		case <-r.done:
			return
		}
	}
}

func (r *RabbitMQ) reconnectLoop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for attempt := 0; ; attempt++ {
		conn, err := amqp.Dial(r.URL)
		if err == nil {
			ch, err := conn.Channel()
			if err == nil {
				r.conn = conn
				r.channel = ch
				r.notifyConnErr = conn.NotifyClose(make(chan *amqp.Error))
				log.Println("[RabbitMQ] Reconnected successfully")
				select {
				case r.ready <- true:
				default:
				}
				return
			}
		}

		backoff := exponentialBackoff(attempt)
		log.Printf("[RabbitMQ] Reconnect failed: %v — retrying in %.1fs", err, backoff.Seconds())
		time.Sleep(backoff)
	}
}

func (r *RabbitMQ) Channel() *amqp.Channel {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.channel
}

func (r *RabbitMQ) NotifyReady() <-chan bool {
	return r.ready
}

func (r *RabbitMQ) Close() {
	r.done <- true
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
	log.Println("[RabbitMQ] Connection closed gracefully")
}

func exponentialBackoff(attempt int) time.Duration {
	baseDelay := 2.0
	maxDelay := 60.0
	delay := baseDelay * math.Pow(2, float64(attempt))
	if delay > maxDelay {
		delay = maxDelay
	}
	return time.Duration(delay) * time.Second
}
