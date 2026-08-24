package main

import (
	"log"

	"github.com/meloop/post-service/messaging"
	"github.com/meloop/services/common/logging"
)

func main() {
	logger := logging.New("post-service")
	logger.Info("service_started")

	err := messaging.PublishPostLiked("user-123", "post-456")
	if err != nil {
		logger.Error("event_publish_failed", "event", "post.liked", "error", err)
		log.Fatal(err)
	}
	logger.Info("event_published", "event", "post.liked")
}
