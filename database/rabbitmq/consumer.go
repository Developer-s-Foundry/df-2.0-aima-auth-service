package rabbitmq

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	WelcomeEmail      = "auth_welcome_mail"
	AuthUser          = "auth_user_info"
	WelcomeEmailQueue = "queue"
)

type Consumer struct {
	rmq        *RabbitMQ
	queueName  string
	consumerID string
	handler    func(amqp.Delivery) error
	done       chan bool
}

func NewConsumer(rmq *RabbitMQ, queueName string, consumerID string, handler func(amqp.Delivery) error) *Consumer {
	return &Consumer{
		rmq:        rmq,
		queueName:  queueName,
		consumerID: consumerID,
		handler:    handler,
		done:       make(chan bool),
	}
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-c.rmq.NotifyReady():
				log.Printf("[Consumer:%s] RabbitMQ connection is ready — starting consumer", c.consumerID)
				c.consume(ctx)
			case <-ctx.Done():
				log.Printf("[Consumer:%s] Context cancelled, shutting down consumer", c.consumerID)
				c.done <- true
				return
			}
		}
	}()
}

func (c *Consumer) consume(ctx context.Context) {
	ch := c.rmq.Channel()
	if ch == nil {
		log.Printf("[Consumer:%s] Channel not ready, waiting...", c.consumerID)
		return
	}

	msgs, err := ch.Consume(
		c.queueName,
		c.consumerID,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("[Consumer:%s] Failed to start consuming: %v", c.consumerID, err)
		return
	}

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("[Consumer:%s] Message channel closed, waiting for reconnect...", c.consumerID)
				return
			}

			if err := c.handler(msg); err != nil {
				log.Printf("[Consumer:%s] Handler error: %v — NACK message", c.consumerID, err)
				_ = msg.Nack(false, true) // requeue
				continue
			}

			_ = msg.Ack(false)
		case <-ctx.Done():
			log.Printf("[Consumer:%s] Context cancelled, stopping message consumption", c.consumerID)
			return
		}
	}
}

func (c *Consumer) Stop() {
	c.done <- true
	log.Printf("[Consumer:%s] Stopped consuming messages", c.consumerID)
}
