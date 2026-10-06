package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/meloop/gamification-service/messaging"
	"github.com/meloop/gamification-service/repositories"
	"github.com/meloop/gamification-service/services"
)

func main() {
	log.Println("Iniciando Gamification Service...")

	// 1. Configurar la conexión al Supabase local de Docker
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		// Fallback directo a la cadena de la guía por si falla la lectura del .env
		connStr = "postgresql://postgres:postgres@127.0.0.1:15422/postgres?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error al abrir la conexión a la base de datos: %v", err)
	}
	defer db.Close()

	// 2. Verificar que la base de datos responde
	if err := db.Ping(); err != nil {
		log.Fatalf("Error al conectar con Supabase local: %v", err)
	}
	log.Println("Conexión a Supabase local establecida con éxito.")

	// 3. Inicializar el Repositorio inyectando la base de datos real
	repo := repositories.NewExperienceRepository(db)

	// 4. Inicializar el Motor de Gamificación y el Servicio
	engine := services.NewGamificationEngine()
	expService := services.NewExperienceService(repo, engine)

	// 5. Configurar y arrancar el consumidor de RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	consumer, err := messaging.NewRabbitMQConsumer(rabbitURL, expService)
	if err != nil {
		log.Fatalf("Error crítico al conectar con RabbitMQ: %v", err)
	}
	defer consumer.Close() // Cierra la conexión de forma segura si la app se apaga

	if err := consumer.StartWorker(); err != nil {
		log.Fatalf("Error al iniciar el worker de RabbitMQ: %v", err)
	}

	// Bloquear el hilo principal para que el worker siga escuchando
	forever := make(chan bool)
	<-forever
}