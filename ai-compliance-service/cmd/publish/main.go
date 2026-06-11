package main

import (
	"flag"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	from := flag.String("from", "alice@company.com", "sender")
	to := flag.String("to", "bob@client.com", "recipient")
	subject := flag.String("subject", "Hello", "subject")
	message := flag.String("message", "Hi there!", "body")
	tenant := flag.String("tenant", "", "tenant_id (optional)")
	flag.Parse()

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel: %v", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare("email_events", "topic", true, false, false, false, nil); err != nil {
		log.Fatalf("declare exchange: %v", err)
	}

	payload := fmt.Sprintf(
		`{"from":%q,"to":%q,"subject":%q,"message":%q,"tenant_id":%q}`,
		*from, *to, *subject, *message, *tenant,
	)

	err = ch.Publish("email_events", "email.ingest", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(payload),
	})
	if err != nil {
		log.Fatalf("publish: %v", err)
	}
	fmt.Printf("Published to email.ingest:\n%s\n", payload)
}
