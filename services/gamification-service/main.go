package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/meloop/gamification-service/messaging"
	"github.com/meloop/gamification-service/repositories"
	"github.com/meloop/gamification-service/services"
)

func main() {
	log.Println("Iniciando Gamification Service...")

	// 1. Inicializar dependencias de la capa de negocio
	repo := repositories.NewExperienceRepository()
	engine := services.NewGamificationEngine()
	expService := services.NewExperienceService(repo, engine)

	// 2. Conectar a RabbitMQ (Asegúrate de que el contenedor en Docker Desktop esté corriendo)
	// La URL por defecto de RabbitMQ local es amqp://guest:guest@localhost:5672/
	rabbitURI := "amqp://meloop:Meloop.67@localhost:5672/"
	consumer, err := messaging.NewRabbitMQConsumer(rabbitURI, expService)
	if err != nil {
		log.Fatalf("Error conectando a RabbitMQ: %v", err)
	}
	defer consumer.Close()

	// 3. Iniciar el worker en segundo plano[cite: 8]
	if err := consumer.StartWorker(); err != nil {
		log.Fatalf("Error iniciando el worker: %v", err)
	}

	// 4. Bloquear el hilo principal para que el microservicio no se cierre instantáneamente
	// Esto permite que el worker reciba eventos en tiempo real de forma constante[cite: 8]
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	log.Println("Apagando Gamification Service de forma segura...")
}