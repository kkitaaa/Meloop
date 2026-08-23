package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/meloop/api-gateway/controllers"
)

func SetupRoutes(router *gin.Engine) {
	router.GET("/health", controllers.Health)

	// Enrutar dinámicamente peticiones bajo /test/* al microservicio de pruebas (recortando prefijo /test)
	router.Any("/test", controllers.ProxyToService("TEST_SERVICE_URL", "http://localhost:8081", "/test"))
	router.Any("/test/*any", controllers.ProxyToService("TEST_SERVICE_URL", "http://localhost:8081", "/test"))

	// Enrutar peticiones de usuarios (conservando el prefijo /users en el destino)
	router.Any("/users", controllers.ProxyToService("USER_SERVICE_URL", "http://localhost:8082", ""))
	router.Any("/users/*any", controllers.ProxyToService("USER_SERVICE_URL", "http://localhost:8082", ""))

	// Enrutar peticiones de autenticación (conservando el prefijo /auth en el destino)
	router.Any("/auth", controllers.ProxyToService("AUTH_SERVICE_URL", "http://localhost:8083", ""))
	router.Any("/auth/*any", controllers.ProxyToService("AUTH_SERVICE_URL", "http://localhost:8083", ""))
}
