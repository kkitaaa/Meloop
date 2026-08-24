package main

import "github.com/meloop/services/common/logging"

func main() {
	logging.New("notification-service").Info("service_started")
}
