package handlers

import (
	"dory-backend/internal/services"
	"dory-backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Chat(c *gin.Context) {
	var input struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Message is required", err.Error())
		return
	}

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

	chunks, err := services.SearchSimilarChunks(userID, input.Message)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Search failed", err.Error())
		return
	}
	aiResponse, err := services.GenerateAIResponse(input.Message, chunks)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Response generated", gin.H{
		"response": aiResponse,
		"sources":  chunks,
	})
}
