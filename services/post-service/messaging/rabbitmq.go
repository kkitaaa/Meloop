package messaging

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/meloop/services/common/logging"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchange     = "meloop.events"
	exchangeType = "topic"
)

type PostLikedEvent struct {
	Event  string `json:"event"`
	UserID string `json:"userId"`
	PostID string `json:"postId"`
}

func PublishPostLiked(userID string, postID string) error {
	logger := logging.New("post-service")
	
	// Leemos la variable de entorno, o usamos tu credencial local por defecto si falla
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://meloop:Meloop.67@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return fmt.Errorf("error conectando a RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("error creando canal: %w", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error declarando exchange: %w", err)
	}

	event := PostLikedEvent{
		Event:  "PostLiked",
		UserID: userID,
		PostID: postID,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	err = ch.Publish(
		exchange,
		"post.liked",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return fmt.Errorf("error publicando evento: %w", err)
	}

	logger.Info("event_published", "event", "post.liked", "payload_bytes", len(body))

	return nil
}