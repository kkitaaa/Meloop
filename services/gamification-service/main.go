package main

import (
	"log"

	"github.com/meloop/gamification-service/messaging"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("gamification-service")
	logger.Info("service_started")

	if err := messaging.StartConsumer(); err != nil {
		logger.Error("event_consumer_failed", "error", err)
		log.Fatal(err)
	}
}
