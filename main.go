package main

import (
	"log"
	"net/http"

	"go-expense-tracker-api/config"
	"go-expense-tracker-api/database"
	_ "go-expense-tracker-api/docs"
	"go-expense-tracker-api/handlers"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Expense Tracker API
// @version 1.0
// @description API documentation for Expense Tracker project.
// @termsOfService http://swagger.io/terms/

// @contact.name Anang
// @contact.url http://github.com/anangkf
// @contact.email gonanggoneng@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
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
	categoryRepo := repositories.NewCategoryRepository(database.DB)

	// INIT HANDLERS
	authHandler := handlers.NewAuthHandler(userRepo, categoryRepo, jwtServices)

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

	// SWAGGER ROUTES
	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// PUBLIC ROUTES
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
}
