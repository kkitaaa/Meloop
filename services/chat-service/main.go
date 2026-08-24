package main

import "github.com/meloop/services/common/logging"

func main() {
	logging.New("chat-service").Info("service_started")
}
