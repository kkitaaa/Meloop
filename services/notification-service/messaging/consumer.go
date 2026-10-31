package messaging

import (
	"context"
	"errors"
	"fmt"

	"github.com/meloop/notification-service/services"
	"github.com/meloop/services/common/logging"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchange           = "meloop.events"
	queueName          = "notification.events"
	deadLetterExchange = "meloop.notifications.dead"
	deadLetterQueue    = "notification.events.dead"
	deadLetterKey      = "notification.dead"
)

var eventRoutingKeys = []string{
	"post.liked",
	"comment.created",
	"comment.commented",
	"comment.replied",
	"friend.requested",
	"friend.accepted",
	"message.sent",
	"report.created",
	"moderation.action",
	"user.level_up",
	"reward.unlocked",
}

func Run(ctx context.Context, rabbitURL string, processor *services.Processor) error {
	logger := logging.New("notification-service")
	connection, err := amqp.Dial(rabbitURL)
	if err != nil {
		return fmt.Errorf("connect to RabbitMQ: %w", err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("open RabbitMQ channel: %w", err)
	}
	defer channel.Close()

	queue, err := declareTopology(channel)
	if err != nil {
		return err
	}
	deliveries, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("start notification consumer: %w", err)
	}
	logger.Info("event_consumer_started", "queue", queue.Name, "bindings", len(eventRoutingKeys))
	return consumeDeliveries(ctx, deliveries, processor, logger)
}

func declareTopology(channel *amqp.Channel) (amqp.Queue, error) {
	if err := channel.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return amqp.Queue{}, fmt.Errorf("declare events exchange: %w", err)
	}
	if err := channel.ExchangeDeclare(deadLetterExchange, "direct", true, false, false, false, nil); err != nil {
		return amqp.Queue{}, fmt.Errorf("declare dead-letter exchange: %w", err)
	}
	deadQueue, err := channel.QueueDeclare(deadLetterQueue, true, false, false, false, nil)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("declare dead-letter queue: %w", err)
	}
	if err := channel.QueueBind(deadQueue.Name, deadLetterKey, deadLetterExchange, false, nil); err != nil {
		return amqp.Queue{}, fmt.Errorf("bind dead-letter queue: %w", err)
	}
	queue, err := channel.QueueDeclare(queueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    deadLetterExchange,
		"x-dead-letter-routing-key": deadLetterKey,
	})
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("declare notification queue: %w", err)
	}
	for _, routingKey := range eventRoutingKeys {
		if err := channel.QueueBind(queue.Name, routingKey, exchange, false, nil); err != nil {
			return amqp.Queue{}, fmt.Errorf("bind event %q: %w", routingKey, err)
		}
	}
	if err := channel.Qos(1, 0, false); err != nil {
		return amqp.Queue{}, fmt.Errorf("configure consumer prefetch: %w", err)
	}
	return queue, nil
}

func consumeDeliveries(ctx context.Context, deliveries <-chan amqp.Delivery, processor *services.Processor, logger interface {
	Error(string, ...any)
}) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("RabbitMQ delivery channel closed")
			}
			if err := handleDelivery(ctx, delivery, processor, logger); err != nil {
				return err
			}
		}
	}
}

func handleDelivery(ctx context.Context, delivery amqp.Delivery, processor *services.Processor, logger interface {
	Error(string, ...any)
}) error {
	if err := processor.Process(ctx, delivery.RoutingKey, delivery.MessageId, delivery.Body); err != nil {
		logger.Error("event_processing_failed", "event", delivery.RoutingKey, "error", err)
		if services.IsPermanent(err) {
			if rejectErr := delivery.Reject(false); rejectErr != nil {
				return fmt.Errorf("reject invalid event: %w", rejectErr)
			}
			return nil
		}
		if nackErr := delivery.Nack(false, true); nackErr != nil {
			return fmt.Errorf("requeue event after processing failure: %w", nackErr)
		}
		return nil
	}
	if err := delivery.Ack(false); err != nil {
		return fmt.Errorf("acknowledge processed event: %w", err)
	}
	return nil
}
