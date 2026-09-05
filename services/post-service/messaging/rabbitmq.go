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

type PostInteractionEvent struct {
	Event  string `json:"event"`
	UserID string `json:"userId"`
	PostID string `json:"postId"`
}

func rabbitURL() string {
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")

	if user == "" {
		user = "guest"
	}

	if password == "" {
		password = "guest"
	}

	return fmt.Sprintf(
		"amqp://%s:%s@localhost:5672/",
		user,
		password,
	)
}

func publishPostEvent(eventName, routingKey, userID, postID string) error {
	logger := logging.New("post-service")

	conn, err := amqp.Dial(rabbitURL())
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

	event := PostInteractionEvent{
		Event:  eventName,
		UserID: userID,
		PostID: postID,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	err = ch.Publish(
		exchange,
		routingKey,
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

	logger.Info(
		"event_published",
		"event", routingKey,
		"payload_bytes", len(body),
	)

	return nil
}

func PublishPostLiked(userID, postID string) error {
	return publishPostEvent(
		"PostLiked",
		"post.liked",
		userID,
		postID,
	)
}

func PublishPostUnliked(userID, postID string) error {
	return publishPostEvent(
		"PostUnliked",
		"post.unliked",
		userID,
		postID,
	)
}
