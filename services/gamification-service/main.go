package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	// 1. Usando localhost para evitar problemas de ruteo de red en Windows
	dbURL := "host=localhost user=postgres password=postgres dbname=postgres port=54322 sslmode=disable"

	// 2. Pasamos explícitamente dbURL al parámetro DSN
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dbURL,
		PreferSimpleProtocol: true, 
	}), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a la base de datos: %v", err)
	}

	// 3. Auto-migrar las nuevas tablas
	err = db.AutoMigrate(&models.UserProgress{}, &models.LevelRule{}, &models.UserReward{})
	if err != nil {
		log.Fatalf("Error migrando la base de datos: %v", err)
	}

	// 4. Sembrar reglas de nivel (Tabla Paramétrica) para poder probar
	seedLevelRules(db)

	// 5. Inyectar dependencias
	repo := repositories.NewGamificationRepository(db)
	gamificationService := services.NewGamificationService(repo)
	gamificationController := controllers.NewGamificationController(gamificationService)

	// 6. Configurar Gin y Rutas
	router := gin.Default()
	routes.SetupRoutes(router, gamificationController)

	// 7. Arrancar servidor en el puerto 8085
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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/gamification-service/config"
	"github.com/meloop/gamification-service/messaging"
	"github.com/meloop/gamification-service/repositories"
	"github.com/meloop/gamification-service/services"
	"github.com/meloop/services/common/logging"
)

func main() {
	if err := run(); err != nil {
		logging.New("gamification-service").Error("service_stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := logging.New("gamification-service")
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("initialize database pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	// 1. Conexión a la base de datos (Supabase Local)
	connStr := "postgresql://postgres:postgres@127.0.0.1:15422/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("db_connection_failed", "error", err)
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("db_ping_failed", "error", err)
	} else {
		logger.Info("db_connected_successfully")
	}

	// 2. Iniciar el consumidor de eventos
	if err := messaging.StartConsumer(); err != nil {
		logger.Error("event_consumer_failed", "error", err)
		log.Fatal(err)
	}
}
	repo := repositories.NewGamificationRepository(pool)
	publisher := messaging.NewPublisher(cfg.RabbitMQURL)
	service := services.NewGamificationService(repo, publisher)

	logger.Info("service_started")
	if err := messaging.StartConsumer(ctx, cfg.RabbitMQURL, service); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("event consumer stopped: %w", err)
	}

	return nil
}
