package main

import (
	"log"
	"net/http"

	"go-expense-tracker-api/config"
	"go-expense-tracker-api/database"

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

	// SETUP ROUTES
	setupRoutes(router)

	return router
}

func setupRoutes(router *gin.Engine /*, authHandler, jwtService*/) {
	// HEALTH CHECK
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK", "message": "Expense Tracker API is running!"})
	})
}
