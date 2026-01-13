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

	// Apply CORS middleware to all routes
	router.Use(middlewares.CORSMiddleware())

	protected := router.Group("/api").Use(middlewares.AuthMiddleware())

	{
		protected.POST("/ingest/pdf", handlers.UploadPDF)
		protected.POST("/ingest/text", handlers.IngestText)
		protected.POST("/chat", middlewares.ExtractUserInfo(), handlers.Chat)
	}

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
