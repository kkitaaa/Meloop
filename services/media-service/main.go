package main

import "github.com/meloop/services/common/logging"

func main() {
	logging.New("media-service").Info("service_started")
}
