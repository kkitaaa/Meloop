package main

import (
	"log"

	"github.com/meloop/gamification-service/messaging"
)

func main() {
	log.Println("Gamification Service iniciado")

	if err := messaging.StartConsumer(); err != nil {
		log.Fatal(err)
	}
}
