package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/meloop/gamification-service/messaging"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("gamification-service")
	logger.Info("service_started")

	// 1. Conexión a la base de datos (Supabase Local)
	connStr := "postgresql://postgres:postgres@127.0.0.1:15422/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("db_connection_failed", "error", err)
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("db_ping_failed", "error", err)
	} else {
		logger.Info("db_connected_successfully")
	}

	// 2. Iniciar el consumidor de eventos
	if err := messaging.StartConsumer(); err != nil {
		logger.Error("event_consumer_failed", "error", err)
		log.Fatal(err)
	}
}