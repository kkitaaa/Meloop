package main

import "github.com/meloop/services/common/logging"

func main() {
	logging.New("social-service").Info("service_started")
}
