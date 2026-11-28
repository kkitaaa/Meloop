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

// UnroutedEventError indica que un mensaje publicado con mandatory=true fue devuelto por RabbitMQ
// al no existir ninguna cola vinculada a la routing key en el exchange.
type UnroutedEventError struct {
	RoutingKey string
	Reason     string
	ReplyCode  uint16
}

func (e *UnroutedEventError) Error() string {
	return fmt.Sprintf("mensaje sin ruta devuelto por RabbitMQ (routing_key: %s, code: %d, reason: %s)", e.RoutingKey, e.ReplyCode, e.Reason)
}

type FriendAcceptedEvent struct {
	EventID    string              `json:"event_id"`
	Event      string              `json:"event"`
	RequestID  models.FriendshipID `json:"request_id"`
	SenderID   string              `json:"sender_id"`
	ReceiverID string              `json:"receiver_id"`
}

type FriendRequestedEvent struct {
	EventID    string              `json:"event_id"`
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

type amqpChannel interface {
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	NotifyReturn(c chan amqp.Return) chan amqp.Return
	Close() error
}

type amqpConnection interface {
	Channel() (amqpChannel, error)
	Close() error
}

type realConnection struct {
	conn *amqp.Connection
}

func (r *realConnection) Channel() (amqpChannel, error) {
	return r.conn.Channel()
}

func (r *realConnection) Close() error {
	return r.conn.Close()
}

type amqpDialer func(url string) (amqpConnection, error)

type Publisher struct {
	url    string
	dialer amqpDialer
}

func NewPublisher(url string) *Publisher {
	return newPublisherWithDialer(url, func(url string) (amqpConnection, error) {
		conn, err := amqp.Dial(url)
		if err != nil {
			return nil, err
		}
		return &realConnection{conn: conn}, nil
	})
}

func newPublisherWithDialer(url string, dialer amqpDialer) *Publisher {
	return &Publisher{url: url, dialer: dialer}
}

func (p *Publisher) PublishFriendAccepted(_ context.Context, request *models.FriendRequest) error {
	eventID, err := newUUID()
	if err != nil {
		return fmt.Errorf("generate friend.accepted event ID: %w", err)
	}
	body, err := json.Marshal(FriendAcceptedEvent{
		EventID: eventID, Event: "friend.accepted", RequestID: request.ID,
		SenderID: request.SenderID, ReceiverID: request.ReceiverID,
	})
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}
	if err := p.publish("friend.accepted", eventID, body, false); err != nil {
		return fmt.Errorf("error publicando evento: %w", err)
	}
	return nil
}

func (p *Publisher) PublishFriendRequested(_ context.Context, request *models.FriendRequest) error {
	if request == nil {
		return fmt.Errorf("friend request is required")
	}
	eventID, err := newUUID()
	if err != nil {
		return fmt.Errorf("generate friend.requested event ID: %w", err)
	}
	body, err := json.Marshal(FriendRequestedEvent{
		EventID: eventID, Event: "friend.requested", RequestID: request.ID,
		SenderID: request.SenderID, ReceiverID: request.ReceiverID,
	})
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}
	if err := p.publish("friend.requested", eventID, body, false); err != nil {
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
	if err := p.publish(friendRequestCancelledRoutingKey, event.EventID, body, true); err != nil {
		return fmt.Errorf("error publicando evento: %w", err)
	}
	return nil
}

func (p *Publisher) publish(routingKey, messageID string, body []byte, mandatory bool) error {
	conn, err := p.dialer(p.url)
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

	var returnChan chan amqp.Return
	if mandatory {
		returnChan = make(chan amqp.Return, 1)
		channel.NotifyReturn(returnChan)
	}

	if err := channel.Publish(exchange, routingKey, mandatory, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    messageID,
		Body:         body,
	}); err != nil {
		return err
	}

	if mandatory {
		closeErr := channel.Close()
		select {
		case ret, ok := <-returnChan:
			if ok {
				logging.New("social-service").Warn("event_unrouted",
					"routing_key", ret.RoutingKey,
					"reason", ret.ReplyText,
					"reply_code", ret.ReplyCode,
					"exchange", ret.Exchange,
				)
				return &UnroutedEventError{
					RoutingKey: ret.RoutingKey,
					Reason:     ret.ReplyText,
					ReplyCode:  ret.ReplyCode,
				}
			}
		default:
		}
		if closeErr != nil {
			return fmt.Errorf("error cerrando canal: %w", closeErr)
		}
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
