module github.com/meloop/post-service

go 1.21

require (
	github.com/meloop/services/common v0.0.0-unpublished
	github.com/rabbitmq/amqp091-go v1.13.0
)

replace github.com/meloop/services/common => ../common
