package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

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
	return p.Publish("notification_exchange", "notification.queue", data)
}

func (p *RabbitMQ) PublishUserManagement(data interface{}) error {
	return p.Publish("user_exchange", "user.queue", data)
}
