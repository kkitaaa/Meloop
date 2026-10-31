package messaging

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	eventsExchange       = "meloop.events"
	activityQueue        = "recommendation.events"
	deadLetterExchange   = "meloop.recommendation.dead"
	deadLetterQueue      = "recommendation.events.dead"
	deadLetterRoutingKey = "recommendation.dead"
)

var activityRoutingKeys = []string{
	"like.created",
	"comment.created",
	"post.created",
}

type ActivityProcessor interface {
	ProcessActivity(ctx context.Context, routingKey, messageID string, body []byte) error
}

func Run(ctx context.Context, rabbitURL string, processor ActivityProcessor) error {
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

	queue, err := declareActivityTopology(channel)
	if err != nil {
		return err
	}
	deliveries, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("start recommendation activity consumer: %w", err)
	}
	slog.Info("recommendation_event_consumer_started", "queue", queue.Name, "bindings", len(activityRoutingKeys))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("RabbitMQ delivery channel closed")
			}
			if err := processDelivery(ctx, delivery, processor); err != nil {
				return err
			}
		}
	}
}

func declareActivityTopology(channel *amqp.Channel) (amqp.Queue, error) {
	if err := channel.ExchangeDeclare(eventsExchange, "topic", true, false, false, false, nil); err != nil {
		return amqp.Queue{}, fmt.Errorf("declare events exchange: %w", err)
	}
	if err := channel.ExchangeDeclare(deadLetterExchange, "direct", true, false, false, false, nil); err != nil {
		return amqp.Queue{}, fmt.Errorf("declare recommendation dead-letter exchange: %w", err)
	}
	deadQueue, err := channel.QueueDeclare(deadLetterQueue, true, false, false, false, nil)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("declare recommendation dead-letter queue: %w", err)
	}
	if err := channel.QueueBind(deadQueue.Name, deadLetterRoutingKey, deadLetterExchange, false, nil); err != nil {
		return amqp.Queue{}, fmt.Errorf("bind recommendation dead-letter queue: %w", err)
	}
	queue, err := channel.QueueDeclare(activityQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    deadLetterExchange,
		"x-dead-letter-routing-key": deadLetterRoutingKey,
	})
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("declare recommendation activity queue: %w", err)
	}
	for _, routingKey := range activityRoutingKeys {
		if err := channel.QueueBind(queue.Name, routingKey, eventsExchange, false, nil); err != nil {
			return amqp.Queue{}, fmt.Errorf("bind recommendation event %q: %w", routingKey, err)
		}
	}
	if err := channel.Qos(1, 0, false); err != nil {
		return amqp.Queue{}, fmt.Errorf("configure recommendation consumer prefetch: %w", err)
	}
	return queue, nil
}

func processDelivery(ctx context.Context, delivery amqp.Delivery, processor ActivityProcessor) error {
	if err := processor.ProcessActivity(ctx, delivery.RoutingKey, delivery.MessageId, delivery.Body); err != nil {
		var permanent interface{ Permanent() bool }
		if errors.As(err, &permanent) && permanent.Permanent() {
			slog.Error("recommendation_event_rejected", "event", delivery.RoutingKey, "error", err)
			if rejectErr := delivery.Reject(false); rejectErr != nil {
				return fmt.Errorf("reject invalid recommendation event: %w", rejectErr)
			}
			return nil
		}
		slog.Error("recommendation_event_processing_failed", "event", delivery.RoutingKey, "error", err)
		if nackErr := delivery.Nack(false, true); nackErr != nil {
			return fmt.Errorf("requeue recommendation event after processing failure: %w", nackErr)
		}
		return nil
	}
	if err := delivery.Ack(false); err != nil {
		return fmt.Errorf("acknowledge processed recommendation event: %w", err)
	}
	return nil
}
