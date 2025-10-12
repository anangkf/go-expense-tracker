package main

import (
	"log"
	"net/http"

	"go-expense-tracker-api/config"
	"go-expense-tracker-api/database"
	"go-expense-tracker-api/handlers"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// LOAD CONFIG
	cfg := config.LoadConfig()

	// DB INIT
	database.InitDatabase(cfg)

	// SETUP GIN MODE
	gin.SetMode(cfg.Server.Mode)

	// SETUP ROUTER
	router := setupRouter(cfg)

	// START SERVER
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// INIT SERVICES
	jwtServices := services.NewJWTService(cfg)

	// INIT REPOSITORIES
	userRepo := repositories.NewUserRepository(database.DB)

	// INIT HANDLERS
	authHandler := handlers.NewAuthHandler(userRepo, jwtServices)

	// SETUP ROUTES
	setupRoutes(router, authHandler, jwtServices)

	return router
}

func setupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, jwtService *services.JWTService) {
	// HEALTH CHECK
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK", "message": "Expense Tracker API is running!"})
	})

	// API v1
	v1 := router.Group("/api/v1")

	// PUBLIC ROUTES
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
	}
}
