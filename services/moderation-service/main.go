package main

import "github.com/meloop/services/common/logging"

func main() {
	logging.New("moderation-service").Info("service_started")
}
