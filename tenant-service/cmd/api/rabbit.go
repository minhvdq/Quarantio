package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	rabbitExchange  = "email_events"
	rabbitIngestKey = "email.ingest"
)

// EmailJob is the message published to RabbitMQ for async compliance checking.
type EmailJob struct {
	From           string `json:"from"`
	To             string `json:"to"`
	Subject        string `json:"subject"`
	Message        string `json:"message"`
	TenantID       string `json:"tenant_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	GmailMessageID string `json:"gmail_message_id,omitempty"`
}

func connectToRabbit(url string) *amqp.Connection {
	if url == "" {
		url = "amqp://guest:guest@rabbitmq"
	}
	var rc int64
	for {
		conn, err := amqp.Dial(url)
		if err != nil {
			log.Println("RabbitMQ not ready, retrying...")
			rc++
		} else {
			log.Println("connected to RabbitMQ")
			return conn
		}
		if rc > 5 {
			log.Println("could not connect to RabbitMQ — async email scanning disabled")
			return nil
		}
		backOff := time.Duration(math.Pow(float64(rc), 2)) * time.Second
		time.Sleep(backOff)
	}
}

func (app *Config) publishEmailJob(ctx context.Context, job EmailJob) error {
	if app.Rabbit == nil {
		return fmt.Errorf("rabbitmq not connected")
	}

	ch, err := app.Rabbit.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(rabbitExchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	return ch.PublishWithContext(ctx,
		rabbitExchange,
		rabbitIngestKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         payload,
		},
	)
}
