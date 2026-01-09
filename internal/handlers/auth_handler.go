package handlers

import (
	"dory-backend/internal/config"
	"dory-backend/internal/models"
	"dory-backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GoogleLogin(c *gin.Context) {
	var input struct {
		IDToken string `json:"idToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idToken is required"})
		return
	}

	payload, err := services.VerfiyGoogleToken(input.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google token"})
		return
	}

	var user models.User
	email := payload.Claims["email"].(string)

	result := config.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		user = models.User{
			Email:        email,
			Name:         payload.Claims["name"].(string),
			ProfilePhoto: payload.Claims["picture"].(string),
			GoogleID:     payload.Subject,
		}
		config.DB.Create(&user)
	}

	// 3. Generate Dory JWT
	token, err := services.GenerateJWTToken(user.ID.String(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})

}
