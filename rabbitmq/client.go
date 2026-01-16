// Package rabbitmq provides RabbitMQ client functionality.
package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/oceanmining/game-manager/config"
)

// Client wraps RabbitMQ connection and channel
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// NewClient creates a new RabbitMQ client
func NewClient(cfg config.RabbitMQConfig) (*Client, error) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.User, cfg.Password, cfg.Host, cfg.Port)

	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("failed to close connection: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue
	_, err = channel.QueueDeclare(
		cfg.Queue, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		if closeErr := channel.Close(); closeErr != nil {
			log.Printf("failed to close channel: %v", closeErr)
		}
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("failed to close connection: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Printf("RabbitMQ connected and queue '%s' declared", cfg.Queue)

	return &Client{
		conn:    conn,
		channel: channel,
		queue:   cfg.Queue,
	}, nil
}

// PublishMatchRequest publishes a match request message to the queue
func (c *Client) PublishMatchRequest(payload []byte) error {
	err := c.channel.Publish(
		"",      // exchange
		c.queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published message to queue '%s'", c.queue)
	return nil
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return err
		}
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
