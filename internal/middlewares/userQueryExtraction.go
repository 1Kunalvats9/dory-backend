package middlewares

import (
	"dory-backend/internal/services"
	"dory-backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ExtractUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Message string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			utils.SendError(c, http.StatusBadRequest, "Message is required", err.Error())
			return
		}

		// Save the message to context for the next handler to use
		c.Set("userMessage", input.Message)

		userIDVal, exists := c.Get("userID")
		if !exists {
			utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "User context missing")
			return
		}

		userID, ok := userIDVal.(string)
		if !ok || userID == "" {
			utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "Invalid user ID")
			return
		}

		info, err := services.RetrieveInfoAndSave(input.Message)

		if err == nil && info != "" {
			_ = services.IngestManualText(userID, info)
		}

		c.Next()
	}
}
