module github.com/meloop/auth-service

go 1.26.6

require (
	github.com/gin-gonic/gin v1.12.0
	github.com/meloop/services/common v0.0.0-unpublished
)

replace github.com/meloop/services/common => ../common
