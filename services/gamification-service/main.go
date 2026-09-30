package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/meloop/gamification-service/controllers"
	"github.com/meloop/gamification-service/models"
	"github.com/meloop/gamification-service/repositories"
	"github.com/meloop/gamification-service/routes"
	"github.com/meloop/gamification-service/services"
)

func main() {
	// 1. Conectar a Supabase Local (PostgreSQL)
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a la base de datos: %v", err)
	}

	// 2. Auto-migrar las nuevas tablas
	err = db.AutoMigrate(&models.UserProgress{}, &models.LevelRule{}, &models.UserReward{})
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

	// 3. Sembrar reglas de nivel (Tabla Paramétrica) para poder probar
	seedLevelRules(db)

	// 4. Inyectar dependencias
	repo := repositories.NewGamificationRepository(db)
	gamificationService := services.NewGamificationService(repo)
	gamificationController := controllers.NewGamificationController(gamificationService)

	// 5. Configurar Gin y Rutas
	router := gin.Default()
	routes.SetupRoutes(router, gamificationController)

	// 6. Arrancar servidor en un puerto diferente (8085)
	log.Println("Gamification Service corriendo en puerto 8085")
	if err := router.Run(":8085"); err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
}

// Función auxiliar para insertar las reglas matemáticas
func seedLevelRules(db *gorm.DB) {
	rules := []models.LevelRule{
		{Level: 2, RequiredXP: 100, RewardName: ""},
		{Level: 3, RequiredXP: 300, RewardName: "Insignia de Bronce"},
		{Level: 4, RequiredXP: 600, RewardName: ""},
		{Level: 5, RequiredXP: 1000, RewardName: "Marco Dorado"},
	}

	for _, rule := range rules {
		db.Where("level = ?", rule.Level).FirstOrCreate(&rule)
	}
}