package main

import "github.com/meloop/services/common/logging"

func main() {
	logging.New("music-service").Info("service_started")
}
