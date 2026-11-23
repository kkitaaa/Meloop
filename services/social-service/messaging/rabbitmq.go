package messaging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/meloop/services/common/logging"
	"github.com/meloop/social-service/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchange                         = "meloop.events"
	exchangeType                     = "topic"
	friendRequestCancelledRoutingKey = "friend.request.cancelled"
)

type FriendAcceptedEvent struct {
	Event      string              `json:"event"`
	RequestID  models.FriendshipID `json:"request_id"`
	SenderID   string              `json:"sender_id"`
	ReceiverID string              `json:"receiver_id"`
}

type FriendRequestCancelledEvent struct {
	EventID    string              `json:"event_id"`
	EventType  string              `json:"event_type"`
	RequestID  models.FriendshipID `json:"request_id"`
	SenderID   string              `json:"sender_id"`
	ReceiverID string              `json:"receiver_id"`
	Reason     string              `json:"reason"`
}

type Publisher struct {
	url string
}

func NewPublisher(url string) *Publisher {
	return &Publisher{url: url}
}

func (p *Publisher) PublishFriendAccepted(_ context.Context, request *models.FriendRequest) error {
	body, err := json.Marshal(FriendAcceptedEvent{
		Event: "friend.accepted", RequestID: request.ID,
		SenderID: request.SenderID, ReceiverID: request.ReceiverID,
	})
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}
	if err := p.publish("friend.accepted", "", body); err != nil {
		return fmt.Errorf("error publicando evento: %w", err)
	}
	return nil
}

func (p *Publisher) PublishFriendRequestCancelled(_ context.Context, request *models.FriendRequest) error {
	event, err := newFriendRequestCancelledEvent(request)
	if err != nil {
		return err
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}
	if err := p.publish(friendRequestCancelledRoutingKey, event.EventID, body); err != nil {
		return fmt.Errorf("error publicando evento: %w", err)
	}
	return nil
}

func (p *Publisher) publish(routingKey, messageID string, body []byte) error {
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
	if err := channel.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		MessageId:   messageID,
		Body:        body,
	}); err != nil {
		return err
	}
	logging.New("social-service").Info("event_published", "event", routingKey, "payload_bytes", len(body))
	return nil
}

func newFriendRequestCancelledEvent(request *models.FriendRequest) (FriendRequestCancelledEvent, error) {
	if request == nil {
		return FriendRequestCancelledEvent{}, fmt.Errorf("friend request is required")
	}
	eventID, err := newUUID()
	if err != nil {
		return FriendRequestCancelledEvent{}, fmt.Errorf("generate event ID: %w", err)
	}
	return FriendRequestCancelledEvent{
		EventID: eventID, EventType: friendRequestCancelledRoutingKey,
		RequestID: request.ID, SenderID: request.SenderID,
		ReceiverID: request.ReceiverID, Reason: "user_blocked",
	}, nil
}

func newUUID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	encoded := hex.EncodeToString(value[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", encoded[:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:]), nil
}
