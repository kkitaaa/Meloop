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

type CommentLikedEvent struct {
	Event     string `json:"event"`
	UserID    string `json:"userId"`
	CommentID string `json:"commentId"`
}

func rabbitURL() string {
	if u := os.Getenv("RABBITMQ_URL"); u != "" {
		return u
	}
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASSWORD")

	if user == "" {
		user = "guest"
	}

	if password == "" {
		password = "guest"
	}

	return fmt.Sprintf("amqp://%s:%s@localhost:5672/", user, password)
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

	err = ch.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil)
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
	return publishPostEvent("PostLiked", "post.liked", userID, postID)
}

func PublishPostUnliked(userID, postID string) error {
	return publishPostEvent("PostUnliked", "post.unliked", userID, postID)
}

func PublishCommentLiked(userID string, commentID string) error {
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

	if err = ch.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("error declarando exchange: %w", err)
	}

	event := CommentLikedEvent{Event: "CommentLiked", UserID: userID, CommentID: commentID}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	if err = ch.Publish(exchange, "like.created", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}); err != nil {
		return fmt.Errorf("error publicando evento comment liked: %w", err)
	}

	logger.Info("event_published", "event", "like.created", "payload_bytes", len(body))
	return nil
}

func PublishCommentUnliked(userID string, commentID string) error {
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

	if err = ch.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("error declarando exchange: %w", err)
	}

	event := CommentLikedEvent{Event: "CommentUnliked", UserID: userID, CommentID: commentID}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	if err = ch.Publish(exchange, "like.deleted", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}); err != nil {
		return fmt.Errorf("error publicando evento comment unliked: %w", err)
	}

	logger.Info("event_published", "event", "like.deleted", "payload_bytes", len(body))
	return nil
}

// CommentCreatedEvent define la estructura del evento (RF-21, RF-22)
type CommentCreatedEvent struct {
	EventID         string `json:"event_id"`
	CommentID       string `json:"comment_id"`
	CommentAuthorID string `json:"comment_author_id"` // Quien comenta (3 XP)
	TargetAuthorID  string `json:"target_author_id"`  // Quien recibe el comentario (4 XP)
	TargetType      string `json:"target_type"`       // "post" o "comment"
	TargetID        string `json:"target_id"`         // ID del post o comentario padre
	CreatedAt       string `json:"fecha"`
}

// PublishCommentCreated publica el evento en RabbitMQ con reintentos básicos (RNF-05)
func PublishCommentCreated(event CommentCreatedEvent) error {
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

	if err = ch.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("error declarando exchange: %w", err)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	if err = ch.Publish(exchange, "comment.created", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // Mensajes persistentes (RNF-05)
		Body:         body,
	}); err != nil {
		return fmt.Errorf("error publicando evento comment.created: %w", err)
	}

	logger.Info("event_published", "event", "comment.created", "payload_bytes", len(body))
	return nil
}
