package main

import (
	"fmt"
	"log"

	"github.com/meloop/post-service/messaging"
)

func main() {
	fmt.Println("Post Service iniciado")

	err := messaging.PublishPostLiked("user-123", "post-456")
	if err != nil {
		log.Fatal(err)
	}
}
