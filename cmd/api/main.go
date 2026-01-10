package main

import (
	"dory-backend/internal/config"
	"dory-backend/internal/handlers"
	"dory-backend/internal/middlewares"
	"dory-backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()      //loading envs
	config.ConnectDatabase() //connecting database
	services.InitQdrant()

	router := gin.Default()
	router.MaxMultipartMemory = 10 << 20

	protected := router.Group("/api").Use(middlewares.AuthMiddleware())

	{
		protected.POST("/ingest/pdf", handlers.UploadPDF)
		protected.POST("/ingest/text", handlers.IngestText)
		protected.POST("/chat", handlers.Chat)
	}

	// CORS preflight handlers for all routes
	router.OPTIONS("/api/auth/google", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Status(204)
	})

	router.OPTIONS("/api/chat", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Status(204)
	})

	router.OPTIONS("/api/ingest/pdf", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Status(204)
	})

	router.OPTIONS("/api/ingest/text", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Status(204)
	})

	router.POST("/api/auth/google", handlers.GoogleLogin)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Dory API",
		})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "the server is running properly",
		})
	})

	router.Run(":" + config.AppConfig.Port)
}
