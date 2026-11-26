package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerUIOrigin = "http://localhost:8090"

func SwaggerUICORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Origin") != swaggerUIOrigin {
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", swaggerUIOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
