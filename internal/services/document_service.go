package services

import (
	"bytes"
	"dory-backend/internal/config"
	"dory-backend/internal/models"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

func ProcessPDF(docID uuid.UUID) {
	go func() {
		var doc models.Document
		if err := config.DB.First(&doc, "id = ?", docID).Error; err != nil {
			log.Printf("Error finding document: %v", err)
			return
		}

		text, err := extractTextFromURL(doc.FileURL)
		if err != nil {
			log.Printf("Extraction failed for %s: %v", doc.Filename, err)
			// Update document status to failed with error details
			config.DB.Model(&doc).Update("status", "failed")
			// Optionally store error message in content field for debugging
			config.DB.Model(&doc).Update("content", fmt.Sprintf("Error: %v", err))
			return
		}

		chunks := ChunkText(text, 300)
		err = StoreChunksInQdrant(doc.UserID.String(), doc.ID.String(), chunks)
		if err != nil {
			log.Printf("Qdrant storage failed: %v", err)
			config.DB.Model(&doc).Update("status", "failed")
			return
		}

		events, err := DetectEvents(text)
		if err == nil {
			for i := range events {
				events[i].UserID = doc.UserID
				events[i].DocumentID = doc.ID
				config.DB.Create(&events[i])
			}
		}

		config.DB.Model(&doc).Updates(models.Document{
			Content: text,
			Status:  "ready",
		})

		log.Printf("Document %s fully processed and embedded", doc.Filename)

		if doc.PublicID != "" {
			DeleteFromCloudinary(doc.PublicID)
			config.DB.Model(&doc).Update("file_url", "")
		}
	}()
}

// Extract text directly from a file stream
func ExtractTextFromFile(file multipart.File) (string, error) {
	// Read entire file into memory
	bodyBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Validate PDF header
	if len(bodyBytes) < 4 || string(bodyBytes[0:4]) != "%PDF" {
		return "", fmt.Errorf("invalid PDF file: file does not start with PDF header")
	}

	bodyReader := bytes.NewReader(bodyBytes)
	pdfReader, err := pdf.NewReader(bodyReader, int64(len(bodyBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %v", err)
	}

	textReader, err := pdfReader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("failed to extract text from PDF: %v", err)
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(textReader)
	if err != nil {
		return "", fmt.Errorf("failed to read extracted text: %v", err)
	}

	text := buf.String()
	if len(text) == 0 {
		return "", fmt.Errorf("PDF appears to be empty or text extraction returned no content")
	}

	return text, nil
}

// Extract text from URL (fallback method)
func extractTextFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download PDF: HTTP %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Validate PDF header
	if len(bodyBytes) < 4 || string(bodyBytes[0:4]) != "%PDF" {
		return "", fmt.Errorf("invalid PDF file: file does not start with PDF header")
	}

	bodyReader := bytes.NewReader(bodyBytes)
	pdfReader, err := pdf.NewReader(bodyReader, int64(len(bodyBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %v", err)
	}

	textReader, err := pdfReader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("failed to extract text from PDF: %v", err)
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(textReader)
	if err != nil {
		return "", fmt.Errorf("failed to read extracted text: %v", err)
	}

	text := buf.String()
	if len(text) == 0 {
		return "", fmt.Errorf("PDF appears to be empty or text extraction returned no content")
	}

	return text, nil
}

func ChunkText(text string, chunkSize int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var chunks []string
	for i := 0; i < len(words); i += chunkSize {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
	}

	return chunks
}
