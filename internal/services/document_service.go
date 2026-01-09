package services

import (
	"bytes"
	"dory-backend/internal/config"
	"dory-backend/internal/models"
	"io"
	"log"
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
			config.DB.Model(&doc).Update("status", "failed")
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
func extractTextFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyReader := bytes.NewReader(bodyBytes)
	pdfReader, err := pdf.NewReader(bodyReader, int64(len(bodyBytes)))
	if err != nil {
		return "", err
	}
	textReader, err := pdfReader.GetPlainText()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(textReader)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
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
