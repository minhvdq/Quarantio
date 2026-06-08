package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type EmailEmitter struct {
	connection *amqp.Connection
}

func NewEmailEmitter(conn *amqp.Connection) (*EmailEmitter, error) {
	emitter := &EmailEmitter{connection: conn}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	if err := declareEmailExchange(ch); err != nil {
		return nil, err
	}

	return emitter, nil
}

func (e *EmailEmitter) Push(payload string, routingKey string) error {
	ch, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.Publish(
		"email_events",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(payload),
		},
	)
}
