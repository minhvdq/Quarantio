package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeName    = "email_events"
	queueName       = "email.compliance.worker"
	ingestKey       = "email.ingest"
	quarantineQueue = "email.quarantine.store"
	quarantineKey   = "email.quarantine"
	blockedQueue    = "email.blocked.store"
	blockedKey      = "email.blocked"
)

type Consumer struct {
	conn *amqp.Connection
}

func NewConsumer(conn *amqp.Connection) (*Consumer, error) {
	c := &Consumer{conn: conn}
	if err := c.setup(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Consumer) setup() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	for _, q := range []struct{ name, key string }{
		{queueName, ingestKey},
		{quarantineQueue, quarantineKey},
		{blockedQueue, blockedKey},
	} {
		if _, err := ch.QueueDeclare(q.name, true, false, false, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(q.name, q.key, exchangeName, false, nil); err != nil {
			return err
		}
	}
	return nil
}

// ConsumeQuarantine returns deliveries from the quarantine queue.
func (c *Consumer) ConsumeQuarantine() (<-chan amqp.Delivery, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		return nil, err
	}
	return ch.Consume(quarantineQueue, "", false, false, false, false, nil)
}

// ConsumeBlocked returns deliveries from the blocked queue.
func (c *Consumer) ConsumeBlocked() (<-chan amqp.Delivery, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		return nil, err
	}
	return ch.Consume(blockedQueue, "", false, false, false, false, nil)
}

// Consume starts delivery on queueName and returns the channel of messages.
func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return nil, err
	}

	return ch.Consume(queueName, "", false, false, false, false, nil)
}
