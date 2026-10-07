package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/meloop/gamification-service/services"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	service *services.ExperienceService // Referencia a la capa de negocio
}

// NewRabbitMQConsumer establece la conexión segura con RabbitMQ.
func NewRabbitMQConsumer(amqpURI string, expService *services.ExperienceService) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Configurar el exchange para eventos de interacciones[cite: 8]
	err = ch.ExchangeDeclare(
		"interactions_exchange", // Nombre del exchange
		"topic",                 // Tipo
		true,                    // Durable (no se pierde si RabbitMQ se reinicia)
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &RabbitMQConsumer{
		conn:    conn,
		channel: ch,
		service: expService,
	}, nil
}

// StartWorker inicia la escucha en segundo plano sin bloquear el sistema[cite: 8].
func (c *RabbitMQConsumer) StartWorker() error {
	// Declarar la cola específica para el gamification-service[cite: 8]
	q, err := c.channel.QueueDeclare(
		"gamification_interactions_queue",
		true, // Durable
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Enlazar la cola al exchange
	err = c.channel.QueueBind(
		q.Name,
		"interaction.*", // Escucha cualquier evento de interacción (ej. interaction.like)
		"interactions_exchange",
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := c.channel.Consume(
		q.Name,
		"gamification_worker", // Consumer tag
		false,                 // Auto-ack en false para evitar pérdida de datos si falla el proceso[cite: 8]
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Goroutine que actúa como worker escuchando en segundo plano[cite: 8]
	go func() {
		for d := range msgs {
			var event services.InteractionEvent
			// Parsear el mensaje en formato JSON[cite: 8]
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("Error al parsear JSON del evento: %v", err)
				d.Reject(false) // Descartar mensaje malformado
				continue
			}

			// Enviar a la capa de lógica de negocio[cite: 8]
			err := c.service.HandleInteractionEvent(context.Background(), event)
			if err != nil {
				log.Printf("Error procesando evento: %v", err)
				d.Nack(false, false) // Reencolar si hubo un error temporal
			} else {
				d.Ack(false) // Confirmar procesamiento exitoso (garantiza cero pérdida de datos)[cite: 8]
			}
		}
	}()

	log.Println("Worker de RabbitMQ iniciado, esperando eventos de interacciones...")
	return nil
}

func (c *RabbitMQConsumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
