package main

import (
	"log"
	"net/http"
	"time"

	"go-expense-tracker-api/config"
	"go-expense-tracker-api/database"
	_ "go-expense-tracker-api/docs"
	"go-expense-tracker-api/handlers"
	"go-expense-tracker-api/middleware"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Expense Tracker API
// @version 1.0
// @description API documentation for Expense Tracker project.
// @termsOfService http://swagger.io/terms/
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @contact.name Anang
// @contact.url http://github.com/anangkf
// @contact.email gonanggoneng@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// // @host localhost:8080
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

	// SETUP CORS MIDDLEWARE
	router.Use(cors.New(cors.Config{
		// AllowOrigins: []string{"http://localhost:3000", "https://your-production-domain.com"},
		AllowAllOrigins:  true, // for development only
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// INIT SERVICES
	jwtServices := services.NewJWTService(cfg)

	// INIT REPOSITORIES
	userRepo := repositories.NewUserRepository(database.DB)
	categoryRepo := repositories.NewCategoryRepository(database.DB)
	expenseRepo := repositories.NewExpenseRepository(database.DB)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(database.DB)

	// INIT HANDLERS
	authHandler := handlers.NewAuthHandler(userRepo, categoryRepo, refreshTokenRepo, jwtServices)
	userHandler := handlers.NewUserHandler(userRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo, userRepo)
	expenseHandler := handlers.NewExpenseHandler(expenseRepo, userRepo, categoryRepo)

	// SETUP ROUTES
	setupRoutes(router, authHandler, userHandler, categoryHandler, expenseHandler, jwtServices)

	return router
}

func setupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, categoryHandler *handlers.CategoryHandler, expenseHandler *handlers.ExpenseHandler, jwtService *services.JWTService) {
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
		auth.POST("/refresh-token", authHandler.RefreshToken)
		auth.POST("/logout", middleware.AuthMiddleware(jwtService), authHandler.Logout)
	}

	// PROTECTED ROUTES
	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware((jwtService)))
	protected.Use(middleware.PaginationAndFilter())
	{
		// USER ROUTES
		user := protected.Group("/user")
		user.GET("/profile", userHandler.GetUserProfile)

		// CATEGORY ROUTES
		category := protected.Group("/categories")
		category.GET("/", categoryHandler.GetCategoriesByUserID)
		category.GET("/default", categoryHandler.GetDefaultCategories)
		category.GET("/:id", categoryHandler.GetCategoryByID)
		category.POST("/", categoryHandler.CreateCategory)
		category.POST("/multiple", categoryHandler.CreateMultipleCategories)
		category.PUT("/:id", categoryHandler.UpdateCategory)
		category.DELETE("/:id", categoryHandler.DeleteCategory)

		// EXPENSE ROUTES
		expense := protected.Group("/expenses")
		expense.GET("/", expenseHandler.GetExpensesByUserID)
		expense.GET("/:id", expenseHandler.GetExpenseByID)
		expense.POST("/", expenseHandler.CreateExpense)
		expense.PUT("/:id", expenseHandler.UpdateExpense)
		expense.DELETE("/:id", expenseHandler.DeleteExpense)
	}
}
