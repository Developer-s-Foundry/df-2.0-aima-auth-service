package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (p *RabbitMQ) DeclareExchangesAndQueues() error {
	ch := p.Channel()
	if ch == nil {
		return fmt.Errorf("RabbitMQ channel not ready")
	}

	exchanges := []string{"notification_exchange", "user_exchange"}
	for _, name := range exchanges {
		if err := ch.ExchangeDeclare(
			name,
			"topic",
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", name, err)
		}
	}

	queues := map[string]string{
		"notification.queue123": "notification_exchange123",
		"user.queue123":         "user_exchange123",
	}

	for queueName, exchangeName := range queues {
		_, err := ch.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}

		if err := ch.QueueBind(
			queueName,
			queueName,
			exchangeName,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w", queueName, exchangeName, err)
		}
	}

	return nil
}

func (p *RabbitMQ) Publish(exchange, routingKey string, data interface{}) error {
	ch := p.Channel()
	if ch == nil {
		return fmt.Errorf("RabbitMQ channel not ready")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	for attempt := 0; attempt < 5; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = ch.PublishWithContext(
			ctx,
			exchange,
			routingKey,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         body,
				DeliveryMode: amqp.Persistent,
				Timestamp:    time.Now(),
			},
		)
		cancel()

		if err == nil {
			log.Printf("[Publisher] Published to %s (key=%s): %s", exchange, routingKey, string(body))
			return nil
		}

		backoff := exponentialBackoff(attempt)
		log.Printf("[Publisher] Publish failed (%v) — retrying in %.1fs", err, backoff.Seconds())
		time.Sleep(backoff)
	}

	return fmt.Errorf("failed to publish message after retries: %w", err)
}

func (p *RabbitMQ) PublishNotification(data interface{}) error {
	return p.Publish("notification_exchange123", "notification123.queue", data)
}

func (p *RabbitMQ) PublishUserManagement(data interface{}) error {
	return p.Publish("user_exchange123", "user123.queue", data)
}
