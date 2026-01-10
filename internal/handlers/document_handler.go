package handlers

import (
	"dory-backend/internal/config"
	"dory-backend/internal/models"
	"dory-backend/internal/services"
	"dory-backend/internal/utils"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

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

	// Validate file extension
	filename := header.Filename
	if !isPDFFile(filename) {
		utils.SendError(c, http.StatusBadRequest, "Invalid file type", "Only PDF files are supported")
		return
	}

	file, err := header.Open()
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to open file", err.Error())
		return
	}
	defer file.Close()

	// Validate PDF file header before uploading
	if !isValidPDF(file) {
		utils.SendError(c, http.StatusBadRequest, "Invalid PDF file", "The file does not appear to be a valid PDF")
		return
	}

	// Reset file pointer after validation
	file.Seek(0, 0)

	// Extract text from PDF before uploading to Cloudinary
	// This avoids the need to download from Cloudinary later (which can have auth issues)
	text, err := services.ExtractTextFromFile(file)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "PDF extraction failed", err.Error())
		return
	}

	// Reset file pointer again for Cloudinary upload
	file.Seek(0, 0)

	cloudURL, publicID, err := services.UploadToCloudinary(file, uuid.New().String())
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Cloud upload failed", err.Error())
		return
	}

	newDoc := models.Document{
		UserID:   userID,
		Filename: filename,
		FileURL:  cloudURL,
		PublicID: publicID,
		Status:   "processing",
		Content:  text, // Store extracted text immediately
	}

	config.DB.Create(&newDoc)

	// Process the extracted text asynchronously (chunking, embedding, event detection)
	go func(docID uuid.UUID, uID uuid.UUID, extractedText string, cloudPublicID string) {
		chunks := services.ChunkText(extractedText, 300)
		err := services.StoreChunksInQdrant(uID.String(), docID.String(), chunks)
		if err != nil {
			log.Printf("Qdrant storage failed for doc %s: %v", docID, err)
			config.DB.Model(&models.Document{}).Where("id = ?", docID).Update("status", "failed")
			return
		}

		events, err := services.DetectEvents(extractedText)
		if err == nil {
			for i := range events {
				events[i].UserID = uID
				events[i].DocumentID = docID
				config.DB.Create(&events[i])
			}
		}

		// Update status to ready
		config.DB.Model(&models.Document{}).Where("id = ?", docID).Update("status", "ready")
		log.Printf("Document %s fully processed and embedded", docID)

		// Delete from Cloudinary after processing (since we have the content stored)
		if cloudPublicID != "" {
			services.DeleteFromCloudinary(cloudPublicID)
			config.DB.Model(&models.Document{}).Where("id = ?", docID).Update("file_url", "")
		}
	}(newDoc.ID, userID, text, publicID)

	utils.SendSuccess(c, http.StatusAccepted, "Upload successful, processing started", newDoc)
}

// Helper function to check if file has PDF extension
func isPDFFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".pdf"
}

// Helper function to validate PDF file header
func isValidPDF(file multipart.File) bool {
	header := make([]byte, 4)
	_, err := file.Read(header)
	if err != nil {
		return false
	}
	// PDF files start with "%PDF"
	return string(header) == "%PDF"
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
