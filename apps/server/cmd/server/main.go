package main

import (
	"net/http"
	"os"
	"outpost-engine/internal/database"
	"outpost-engine/internal/handlers"
	"outpost-engine/internal/middleware"
	"outpost-engine/internal/modules/post"
	"outpost-engine/internal/modules/user"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var logger *zap.Logger

func initLogger() {
	var err error
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
}

func main() {
	// Load environment variables
	godotenv.Load()

	// Initialize logger
	initLogger()
	defer logger.Sync()

	logger.Info("Starting outpost-engine server")

	db, err := database.ConnectDatabase(logger)
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
	}
	defer db.Close()

	// Get host from environment
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	// Get HTTP port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	addr := host + ":" + port

	logger.Info("Starting HTTP server", zap.String("address", addr))

	// CORS: pinned to CORS_ORIGIN when set, permissive otherwise.
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}

	// Create Gin router
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Asynq client (Redis) — enqueues publish tasks
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqClient.Close()

	// handler post, social, dan user
	userHandler := user.NewUserHandler(db)
	postHandler := post.NewPostHandler(db, asynqClient)

	// Public auth routes
	r.POST("/auth/register", userHandler.Register)
	r.POST("/auth/login", userHandler.Login)

	// Protected routes
	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/auth/me", userHandler.Me) // profile

	auth.POST("/posts", postHandler.CreatePost)
	auth.GET("/posts", postHandler.GetPosts)

	// Health check endpoint
	r.GET("/health", handlers.HealthCheck)

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to outpost-engine!",
		})
	})

	// Start server
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
