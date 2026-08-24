package main

import "github.com/meloop/services/common/logging"

func main() {
	logging.New("recommendation-service").Info("service_started")
}
