package messaging

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	rabbitURL    = "amqp://guest:guest@localhost:5672/"
	exchange     = "meloop.events"
	exchangeType = "topic"
)

type PostLikedEvent struct {
	Event  string `json:"event"`
	UserID string `json:"userId"`
	PostID string `json:"postId"`
}

func PublishPostLiked(userID string, postID string) error {
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

	fmt.Println("Evento PostLiked publicado:", string(body))

	return nil
}
