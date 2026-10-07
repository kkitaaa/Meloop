package messaging

import (
	"encoding/json"
	"fmt"

	"github.com/meloop/services/common/logging"
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

type CommentLikedEvent struct {
	Event     string `json:"event"`
	UserID    string `json:"userId"`
	CommentID string `json:"commentId"`
}

// Funciones para Posts (de la tarea anterior)
func PublishPostLiked(userID string, postID string) error {
	logger := logging.New("post-service")
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

	err = ch.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error declarando exchange: %w", err)
	}

	event := PostLikedEvent{Event: "PostLiked", UserID: userID, PostID: postID}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	err = ch.Publish(exchange, "post.liked", false, false, amqp.Publishing{ContentType: "application/json", Body: body})
	if err != nil {
		return fmt.Errorf("error publicando evento: %w", err)
	}

	logger.Info("event_published", "event", "post.liked", "payload_bytes", len(body))
	return nil
}

// Funciones para Comentarios (Nueva tarea)
func PublishCommentLiked(userID string, commentID string) error {
	logger := logging.New("post-service")
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

	err = ch.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error declarando exchange: %w", err)
	}

	event := CommentLikedEvent{Event: "CommentLiked", UserID: userID, CommentID: commentID}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	err = ch.Publish(exchange, "like.created", false, false, amqp.Publishing{ContentType: "application/json", Body: body})
	if err != nil {
		return fmt.Errorf("error publicando evento comment liked: %w", err)
	}

	logger.Info("event_published", "event", "like.created", "payload_bytes", len(body))
	return nil
}

func PublishCommentUnliked(userID string, commentID string) error {
	logger := logging.New("post-service")
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

	err = ch.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error declarando exchange: %w", err)
	}

	event := CommentLikedEvent{Event: "CommentUnliked", UserID: userID, CommentID: commentID}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	err = ch.Publish(exchange, "like.deleted", false, false, amqp.Publishing{ContentType: "application/json", Body: body})
	if err != nil {
		return fmt.Errorf("error publicando evento comment unliked: %w", err)
	}

	logger.Info("event_published", "event", "like.deleted", "payload_bytes", len(body))
	return nil
}