package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blog-api/internal/config"
	"blog-api/internal/handlers"
	"blog-api/internal/infrastructure"
	"blog-api/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize infrastructure manager
	infraManager, err := infrastructure.NewManager(cfg)
	if err != nil {
		log.Fatal("Failed to initialize infrastructure manager:", err)
	}
	defer infraManager.Close()

	// Initialize services
	postService := services.NewPostService(infraManager)

	// Initialize handlers
	postHandler := handlers.NewPostHandler(postService)

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Post endpoints
		api.POST("/posts", postHandler.CreatePost)
		api.GET("/posts/:id", postHandler.GetPost)
		api.PUT("/posts/:id", postHandler.UpdatePost)
		api.GET("/posts/search-by-tag", postHandler.SearchByTag)
		api.GET("/posts/search", postHandler.SearchPosts)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		health := infraManager.Health()
		isHealthy := infraManager.IsHealthy()

		status := http.StatusOK
		if !isHealthy {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status":    "healthy",
			"time":      time.Now().Format(time.RFC3339),
			"services":  health,
			"stats":     infraManager.Stats(),
		})
	})

	// Setup HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
