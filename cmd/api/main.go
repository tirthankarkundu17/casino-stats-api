package main

import (
	"tirthankarkundu17/casino-analytics/internal/cache"
	"tirthankarkundu17/casino-analytics/internal/config"
	"tirthankarkundu17/casino-analytics/internal/db"
	"tirthankarkundu17/casino-analytics/internal/handlers"
	"tirthankarkundu17/casino-analytics/internal/logger"
	"tirthankarkundu17/casino-analytics/internal/repository"
	"tirthankarkundu17/casino-analytics/internal/routes"
	"tirthankarkundu17/casino-analytics/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Logger
	log := logger.New()
	defer log.Sync()

	// Load Configuration
	cfg := config.LoadConfig(log)

	// Initialize Mongo
	m, err := db.ConnectMongo(cfg.MongoURI, cfg.MongoDB, log)
	if err != nil {
		log.Fatalf("Failed to connect to Mongo: %v", err)
	}
	defer m.Close()

	// Initialize Cache
	c := cache.NewCache(cfg.RedisAddr)

	// Initialize Repository, Services & Handlers
	repo := repository.NewTransactionRepository(m.Db, cfg.TransactionCollection, log)
	analyticsService := services.NewAnalyticsService(repo)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, c, log)

	// Setup Router
	r := gin.Default()
	routes.SetupRoutes(r, analyticsHandler, cfg.StaticAuthToken)

	log.Infof("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
