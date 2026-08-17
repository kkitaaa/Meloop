package messaging

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	rabbitURL    = "amqp://guest:guest@localhost:5672/"
	exchange     = "meloop.events"
	exchangeType = "topic"
	queueName    = "gamification.events"
	routingKey   = "post.liked"
)

func StartConsumer() error {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return fmt.Errorf("error conectando a RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("error creando canal: %w", err)
	}

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
		ch.Close()
		conn.Close()
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
		ch.Close()
		conn.Close()
		return fmt.Errorf("error declarando cola: %w", err)
	}

	err = ch.QueueBind(
		queue.Name,
		routingKey,
		exchange,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("error vinculando cola: %w", err)
	}

	messages, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("error creando consumidor: %w", err)
	}

	fmt.Println("Gamification Service escuchando eventos...")

	for message := range messages {
		fmt.Printf(
			"Evento recibido: %s\n",
			string(message.Body),
		)
	}

	return nil
}
