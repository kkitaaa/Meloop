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
	router.Any("/users", controllers.SessionAuth(), controllers.ProxyToService("USER_SERVICE_URL", "http://localhost:8082", ""))
	router.Any("/users/*any", controllers.SessionAuth(), controllers.ProxyToService("USER_SERVICE_URL", "http://localhost:8082", ""))

	// Enrutar peticiones de autenticación (conservando el prefijo /auth en el destino)
	router.Any("/auth", controllers.ProxyToService("AUTH_SERVICE_URL", "http://localhost:8083", ""))
	router.Any("/auth/*any", controllers.ProxyToService("AUTH_SERVICE_URL", "http://localhost:8083", ""))

	// Enrutar peticiones de relaciones sociales (conservando el prefijo /friends en el destino)
	router.Any("/friends", controllers.SessionAuth(), controllers.ProxyToService("SOCIAL_SERVICE_URL", "http://localhost:8086", ""))
	router.Any("/friends/*any", controllers.SessionAuth(), controllers.ProxyToService("SOCIAL_SERVICE_URL", "http://localhost:8086", ""))

	// Enrutar peticiones de recomendaciones (conservando el prefijo /recommendations en el destino)
	router.Any("/recommendations", controllers.SessionAuth(), controllers.ProxyToService("RECOMMENDATION_SERVICE_URL", "http://localhost:8087", ""))
	router.Any("/recommendations/*any", controllers.SessionAuth(), controllers.ProxyToService("RECOMMENDATION_SERVICE_URL", "http://localhost:8087", ""))

	// Enrutar peticiones multimedia (conservando el prefijo /media en el destino)
	router.Any("/media", controllers.SessionAuth(), controllers.ProxyToService("MEDIA_SERVICE_URL", "http://localhost:8085", ""))
	router.Any("/media/*any", controllers.SessionAuth(), controllers.ProxyToService("MEDIA_SERVICE_URL", "http://localhost:8085", ""))

	// Enrutar notificaciones (conservando el prefijo /notifications en el destino)
	router.Any("/notifications", controllers.SessionAuth(), controllers.ProxyToService("NOTIFICATION_SERVICE_URL", "http://localhost:8088", ""))
	router.Any("/notifications/*any", controllers.SessionAuth(), controllers.ProxyToService("NOTIFICATION_SERVICE_URL", "http://localhost:8088", ""))

	// Enrutar peticiones versionadas v1
	router.Any("/v1/users", controllers.ProxyToService("USER_SERVICE_URL", "http://localhost:8082", ""))
	router.Any("/v1/users/*any", controllers.ProxyToService("USER_SERVICE_URL", "http://localhost:8082", ""))
	router.Any("/v1/social/*any", controllers.ProxyToService("SOCIAL_SERVICE_URL", "http://localhost:8086", ""))
	router.Any("/v1/media", controllers.ProxyToService("MEDIA_SERVICE_URL", "http://localhost:8085", ""))
	router.Any("/v1/media/*any", controllers.ProxyToService("MEDIA_SERVICE_URL", "http://localhost:8085", ""))

	// Enrutar peticiones de Post Service
	for _, prefix := range []string{
		"/api/v1/posts",
		"/v1/posts",
		"/v1/feed",
		"/posts",
	} {
		router.Any(prefix,
			controllers.SessionAuth(),
			controllers.ProxyToService("POST_SERVICE_URL", "http://localhost:8084", ""),
		)
		router.Any(prefix+"/*any",
			controllers.SessionAuth(),
			controllers.ProxyToService("POST_SERVICE_URL", "http://localhost:8084", ""),
		)
	}

	// Enrutar peticiones de gamificación
	router.Any("/api/v1/gamification",
		controllers.ProxyToService(
			"GAMIFICATION_SERVICE_URL",
			"http://gamification-service.student-amanda-garcia-4679dc667f08.svc.cluster.local:8089",
			"",
		),
	)

	router.Any("/api/v1/gamification/*any",
		controllers.ProxyToService(
			"GAMIFICATION_SERVICE_URL",
			"http://gamification-service.student-amanda-garcia-4679dc667f08.svc.cluster.local:8089",
			"",
		),
	)

	// Enrutar peticiones de publicaciones, comentarios, respuestas, likes y feed principal
	postServiceDefault := "http://post-service.student-mcrisosto.svc.cluster.local:8084"
	router.Any("/api/v1/posts", controllers.SessionAuth(), controllers.ProxyToService("POST_SERVICE_URL", postServiceDefault, ""))
	router.Any("/api/v1/posts/*any", controllers.SessionAuth(), controllers.ProxyToService("POST_SERVICE_URL", postServiceDefault, ""))
	router.Any("/v1/feed", controllers.SessionAuth(), controllers.ProxyToService("POST_SERVICE_URL", postServiceDefault, ""))
	router.Any("/v1/feed/*any", controllers.SessionAuth(), controllers.ProxyToService("POST_SERVICE_URL", postServiceDefault, ""))
	router.Any("/v1/posts", controllers.SessionAuth(), controllers.ProxyToService("POST_SERVICE_URL", postServiceDefault, ""))
	router.Any("/v1/posts/*any", controllers.SessionAuth(), controllers.ProxyToService("POST_SERVICE_URL", postServiceDefault, ""))
	router.Any("/posts", controllers.SessionAuth(), controllers.ProxyToService("POST_SERVICE_URL", postServiceDefault, ""))
	router.Any("/posts/*any", controllers.SessionAuth(), controllers.ProxyToService("POST_SERVICE_URL", postServiceDefault, ""))
}
