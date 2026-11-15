package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/meloop/gamification-service/models"
	"github.com/meloop/services/common/logging"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	defaultRabbitURL         = "amqp://guest:guest@localhost:5672/"
	exchange                 = "meloop.events"
	exchangeType             = "topic"
	queueName                = "gamification.events"
	RoutingKeyLevelUp        = "level.up"
	RoutingKeyRewardUnlocked = "reward.unlocked"
)

var consumerRoutingKeys = []string{
	"post.liked",
	"comment.created",
}

// ExperienceProcessor es la interfaz utilizada por el consumidor para procesar XP
type ExperienceProcessor interface {
	ProcessExperience(ctx context.Context, userID string, amount int) (*models.ExperienceResult, error)
}

// Publisher maneja la publicación de eventos RabbitMQ
type Publisher struct {
	url string
}

func NewPublisher(url string) *Publisher {
	if strings.TrimSpace(url) == "" {
		url = defaultRabbitURL
	}
	return &Publisher{url: url}
}

func (p *Publisher) PublishLevelUp(ctx context.Context, event models.LevelUpEvent) error {
	if event.Event == "" {
		event.Event = RoutingKeyLevelUp
	}
	if event.Level == 0 && event.Nivel > 0 {
		event.Level = event.Nivel
	}
	if event.Nivel == 0 && event.Level > 0 {
		event.Nivel = event.Level
	}
	if err := p.publish(ctx, RoutingKeyLevelUp, event); err != nil {
		return err
	}
	// Publicar también a user.level_up para compatibilidad con servicios existentes
	_ = p.publish(ctx, "user.level_up", event)
	return nil
}

func (p *Publisher) PublishRewardUnlocked(ctx context.Context, event models.RewardUnlockedEvent) error {
	if event.Event == "" {
		event.Event = RoutingKeyRewardUnlocked
	}
	if event.Level == 0 && event.Nivel > 0 {
		event.Level = event.Nivel
	}
	if event.Nivel == 0 && event.Level > 0 {
		event.Nivel = event.Level
	}
	return p.publish(ctx, RoutingKeyRewardUnlocked, event)
}

func (p *Publisher) publish(ctx context.Context, routingKey string, payload any) error {
	conn, err := amqp.Dial(p.url)
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

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	err = ch.PublishWithContext(
		ctx,
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

	logging.New("gamification-service").Info("event_published", "event", routingKey, "payload_bytes", len(body))
	return nil
}

// StartConsumer inicia el consumo de eventos para el gamification-service
func StartConsumer(ctx context.Context, rabbitURL string, processor ExperienceProcessor) error {
	logger := logging.New("gamification-service")
	if strings.TrimSpace(rabbitURL) == "" {
		rabbitURL = defaultRabbitURL
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

	queue, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error declarando cola: %w", err)
	}

	for _, rk := range consumerRoutingKeys {
		if err := ch.QueueBind(queue.Name, rk, exchange, false, nil); err != nil {
			return fmt.Errorf("error vinculando cola para %s: %w", rk, err)
		}
	}

	messages, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error creando consumidor: %w", err)
	}

	logger.Info("event_consumer_started", "queue", queueName, "bindings", len(consumerRoutingKeys))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-messages:
			if !ok {
				return fmt.Errorf("canal de mensajes cerrado")
			}
			logger.Info("event_consumed", "event", msg.RoutingKey, "payload_bytes", len(msg.Body))

			if err := handleConsumedMessage(ctx, msg.RoutingKey, msg.Body, processor); err != nil {
				logger.Error("event_consume_failed", "event", msg.RoutingKey, "error", err)
				_ = msg.Nack(false, false)
				continue
			}

			_ = msg.Ack(false)
		}
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func handleConsumedMessage(ctx context.Context, routingKey string, body []byte, processor ExperienceProcessor) error {
	var payload struct {
		ActorID         string `json:"actor_id"`
		CommentAuthorID string `json:"comment_author_id"`
		UserID          string `json:"userId"`
		UserIDSnake     string `json:"user_id"`
		TargetOwnerID   string `json:"target_owner_id"`
		TargetAuthorID  string `json:"target_author_id"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("error deserializando evento %s: %w", routingKey, err)
	}

	actorID := firstNonEmpty(payload.ActorID, payload.CommentAuthorID, payload.UserIDSnake, payload.UserID)
	if actorID == "" {
		return fmt.Errorf("evento %s sin identificador de actor/usuario válido", routingKey)
	}

	if processor == nil {
		return nil
	}

	if routingKey == "comment.created" {
		targetOwnerID := firstNonEmpty(payload.TargetOwnerID, payload.TargetAuthorID)

		// Regla anti-autofarmeo (RN-20): no se otorga XP por comentar contenido propio
		if targetOwnerID != "" && actorID == targetOwnerID {
			logging.New("gamification-service").Info("self_comment_ignored", "actor_id", actorID)
			return nil
		}

		// +3 XP para quien comenta (RF-21/RF-22)
		if _, err := processor.ProcessExperience(ctx, actorID, 3); err != nil {
			return fmt.Errorf("error asignando XP al autor del comentario: %w", err)
		}

		// +4 XP para el autor de la publicación o comentario padre
		if targetOwnerID != "" {
			if _, err := processor.ProcessExperience(ctx, targetOwnerID, 4); err != nil {
				return fmt.Errorf("error asignando XP al destinatario del comentario: %w", err)
			}
		}
		return nil
	}

	// Eventos de interacción simples (ej. post.liked)
	_, err := processor.ProcessExperience(ctx, actorID, 10)
	return err
}
