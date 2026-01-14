package handlers

import (
	"dory-backend/internal/services"
	"dory-backend/internal/utils"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
)

func Chat(c *gin.Context) {
	var input struct {
		Message string `json:"message" binding:"required"`
	}

	// Try to get the message from context first (populated by middleware)
	if msg, exists := c.Get("userMessage"); exists {
		if m, ok := msg.(string); ok {
			input.Message = m
		}
	}

	// If not in context, bind from JSON
	if input.Message == "" {
		if err := c.ShouldBindJSON(&input); err != nil {
			utils.SendError(c, http.StatusBadRequest, "Message is required", err.Error())
			return
		}
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

func ChatStream(c *gin.Context) {
	var input struct {
		Message string `json:"message" binding:"required"`
	}

	if msg, exists := c.Get("userMessage"); exists {
		if m, ok := msg.(string); ok {
			input.Message = m
		}
	}

	if input.Message == "" {
		if err := c.ShouldBindJSON(&input); err != nil {
			utils.SendError(c, http.StatusBadRequest, "Message is required", err.Error())
			return
		}
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "User context missing")
		return
	}

	userID := userIDVal.(string)

	chunks, err := services.SearchSimilarChunks(userID, input.Message)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Search failed", err.Error())
		return
	}

	iterRaw, err := services.StreamAIResponse(c.Request.Context(), input.Message, chunks)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}

	iter, ok := iterRaw.(interface {
		Next() (*genai.GenerateContentResponse, error)
	})
	if !ok {
		utils.SendError(c, http.StatusInternalServerError, "AI generation failed", "Invalid iterator type")
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	c.Stream(func(w io.Writer) bool {
		resp, err := iter.Next()
		if err == iterator.Done {
			return false
		}
		if err != nil {
			return false
		}

		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			part := resp.Candidates[0].Content.Parts[0]
			c.SSEvent("message", gin.H{
				"chunk": fmt.Sprintf("%v", part),
			})
		}
		return true
	})
}
