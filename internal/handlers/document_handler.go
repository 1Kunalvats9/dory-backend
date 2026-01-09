package handlers

import (
	"dory-backend/internal/config"
	"dory-backend/internal/models"
	"dory-backend/internal/services"
	"dory-backend/internal/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Helper to get UserID safely to avoid repeating code
func getAuthUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	uidStr, ok := val.(string)
	if !ok {
		return uuid.Nil, false
	}
	uid, err := uuid.Parse(uidStr)
	return uid, err == nil
}

func UploadPDF(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "User context missing")
		return
	}

	header, err := c.FormFile("file")
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "File is required", err.Error())
		return
	}

	file, err := header.Open()
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to open file", err.Error())
		return
	}
	defer file.Close()

	cloudURL, publicID, err := services.UploadToCloudinary(file, uuid.New().String())
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Cloud upload failed", err.Error())
		return
	}

	newDoc := models.Document{
		UserID:   userID,
		Filename: header.Filename,
		FileURL:  cloudURL,
		PublicID: publicID,
		Status:   "processing",
	}

	config.DB.Create(&newDoc)

	// Start background processing
	services.ProcessPDF(newDoc.ID)

	utils.SendSuccess(c, http.StatusAccepted, "Upload successful, processing started", newDoc)
}

func IngestText(c *gin.Context) {
	userID, ok := getAuthUserID(c)
	if !ok {
		utils.SendError(c, http.StatusUnauthorized, "Unauthorized", "User context missing")
		return
	}

	var input struct {
		Content  string `json:"content" binding:"required"`
		Filename string `json:"filename"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Content is required", err.Error())
		return
	}

	newDoc := models.Document{
		UserID:   userID,
		Filename: input.Filename,
		Content:  input.Content,
		FileType: "text",
		Status:   "ready",
	}

	if newDoc.Filename == "" {
		newDoc.Filename = "Quick Note " + uuid.New().String()[:8]
	}

	config.DB.Create(&newDoc)

	// Async processing for Qdrant and Event Detection
	go func(uID uuid.UUID, docID uuid.UUID, content string) {
		chunks := services.ChunkText(content, 300)
		err := services.StoreChunksInQdrant(uID.String(), docID.String(), chunks)
		if err != nil {
			log.Printf("Text embedding failed for doc %s: %v", docID, err)
			config.DB.Model(&models.Document{}).Where("id = ?", docID).Update("status", "failed")
			return
		}

		events, err := services.DetectEvents(content)
		if err == nil {
			for i := range events {
				events[i].UserID = uID
				events[i].DocumentID = docID
				config.DB.Create(&events[i])
			}
		}
	}(userID, newDoc.ID, input.Content) // Pass variables explicitly to the closure

	utils.SendSuccess(c, http.StatusCreated, "Text ingested and embedding started", newDoc)
}
