package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

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

// CommentCreatedEvent define el contrato unificado entre post-service y gamification-service (RF-21, RF-22, RN-20)
type CommentCreatedEvent struct {
	EventID         string `json:"event_id"`
	Event           string `json:"event"`
	ActionType      string `json:"action_type"` // "COMMENT"
	CommentID       string `json:"comment_id"`
	ActorID         string `json:"actor_id"`          // Quien comenta (estándar GamificationEngine)
	UserID          string `json:"user_id"`           // Compatibilidad consumidor
	CommentAuthorID string `json:"comment_author_id"` // Quien comenta (3 XP)
	TargetOwnerID   string `json:"target_owner_id"`   // Dueño del post/comentario (estándar GamificationEngine)
	TargetAuthorID  string `json:"target_author_id"`  // Quien recibe el comentario (4 XP)
	TargetType      string `json:"target_type"`       // "post" o "comment"
	TargetID        string `json:"target_id"`         // ID del post o comentario padre
	Timestamp       string `json:"timestamp"`         // RFC3339
	CreatedAt       string `json:"fecha"`             // Compatibilidad legacy
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

func publishWithConfirm(routingKey string, body []byte) error {
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

	// Habilitar Publisher Confirms para garantizar entrega al broker (RNF-05)
	if err := ch.Confirm(false); err != nil {
		return fmt.Errorf("error activando confirmaciones del broker: %w", err)
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	if err := ch.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("error declarando exchange: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("error publicando evento %s: %w", routingKey, err)
	}

	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("RabbitMQ rechazó (NACK) el mensaje %s", routingKey)
		}
		return nil
	case <-time.After(3 * time.Second):
		return fmt.Errorf("timeout esperando confirmación (ACK) de RabbitMQ para %s", routingKey)
	}
}

// PublishCommentCreated publica el evento con confirmaciones del broker y reintentos con backoff exponencial (RNF-05)
func PublishCommentCreated(event CommentCreatedEvent) error {
	logger := logging.New("post-service")

	if event.Event == "" {
		event.Event = "comment.created"
	}
	if event.ActionType == "" {
		event.ActionType = "COMMENT"
	}
	if event.ActorID == "" {
		event.ActorID = event.CommentAuthorID
	}
	if event.UserID == "" {
		event.UserID = event.CommentAuthorID
	}
	if event.TargetOwnerID == "" {
		event.TargetOwnerID = event.TargetAuthorID
	}
	if event.Timestamp == "" {
		event.Timestamp = event.CreatedAt
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	maxRetries := 3
	backoff := 150 * time.Millisecond
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := publishWithConfirm("comment.created", body); err != nil {
			lastErr = err
			logger.Error("event_publish_retry", "event", "comment.created", "attempt", attempt, "error", err)
			if attempt < maxRetries {
				time.Sleep(backoff)
				backoff *= 2
			}
			continue
		}

		logger.Info("event_published", "event", "comment.created", "attempt", attempt, "payload_bytes", len(body))
		return nil
	}

	return fmt.Errorf("falló la entrega confirmada a RabbitMQ tras %d intentos: %w", maxRetries, lastErr)
}

func publishPostEvent(eventName, routingKey, userID, postID string) error {
	logger := logging.New("post-service")
	event := PostInteractionEvent{
		Event:  eventName,
		UserID: userID,
		PostID: postID,
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}
	if err := publishWithConfirm(routingKey, body); err != nil {
		return err
	}
	logger.Info("event_published", "event", routingKey, "payload_bytes", len(body))
	return nil
}

func PublishPostCreated(authorID, postID string) error {
	return publishPostEvent("PostCreated", "post.created", authorID, postID)
}

func PublishPostLiked(userID, postID string) error {
	return publishPostEvent("PostLiked", "post.liked", userID, postID)
}

func PublishPostUnliked(userID, postID string) error {
	return publishPostEvent("PostUnliked", "post.unliked", userID, postID)
}

func PublishCommentLiked(userID string, commentID string) error {
	logger := logging.New("post-service")
	event := CommentLikedEvent{Event: "CommentLiked", UserID: userID, CommentID: commentID}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}
	if err := publishWithConfirm("like.created", body); err != nil {
		return err
	}
	logger.Info("event_published", "event", "like.created", "payload_bytes", len(body))
	return nil
}

func PublishCommentUnliked(userID string, commentID string) error {
	logger := logging.New("post-service")
	event := CommentLikedEvent{Event: "CommentUnliked", UserID: userID, CommentID: commentID}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}
	if err := publishWithConfirm("like.deleted", body); err != nil {
		return err
	}
	logger.Info("event_published", "event", "like.deleted", "payload_bytes", len(body))
	return nil
}
