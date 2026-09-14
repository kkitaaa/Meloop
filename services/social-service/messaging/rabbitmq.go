package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/meloop/services/common/logging"
	"github.com/meloop/social-service/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchange     = "meloop.events"
	exchangeType = "topic"
)

type FriendAcceptedEvent struct {
	Event      string `json:"event"`
	RequestID  int    `json:"request_id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
}

type Publisher struct {
	url string
}

func NewPublisher(url string) *Publisher {
	return &Publisher{url: url}
}

func (p *Publisher) PublishFriendAccepted(_ context.Context, request *models.FriendRequest) error {
	conn, err := amqp.Dial(p.url)
	if err != nil {
		return fmt.Errorf("error conectando a RabbitMQ: %w", err)
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("error creando canal: %w", err)
	}
	defer channel.Close()
	if err := channel.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("error declarando exchange: %w", err)
	}
	body, err := json.Marshal(FriendAcceptedEvent{
		Event: "friend.accepted", RequestID: request.ID,
		SenderID: request.SenderID, ReceiverID: request.ReceiverID,
	})
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}
	if err := channel.Publish(exchange, "friend.accepted", false, false, amqp.Publishing{ContentType: "application/json", Body: body}); err != nil {
		return fmt.Errorf("error publicando evento: %w", err)
	}
	logging.New("social-service").Info("event_published", "event", "friend.accepted", "payload_bytes", len(body))
	return nil
}
