package handlers

import (
	"dory-backend/internal/config"
	"dory-backend/internal/models"
	"dory-backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetDetectedEvents(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "User context missing")
		return
	}

	var events []models.DetectedEvent
	if err := config.DB.Where("user_id = ?", userID).Order("detected_at DESC").Find(&events).Error; err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch events", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Events retrieved successfully", events)
}

func GetUpcomingEvents(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "User context missing")
		return
	}

	var events []models.DetectedEvent
	// Filter for events where StartTime is in the future
	if err := config.DB.Where("user_id = ? AND start_time > ?", userID, "NOW()").Order("start_time ASC").Find(&events).Error; err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch upcoming events", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Upcoming events retrieved successfully", events)
}
