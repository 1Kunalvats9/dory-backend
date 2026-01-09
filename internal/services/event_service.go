package services

import (
	"context"
	"dory-backend/internal/config"
	"dory-backend/internal/models"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func DetectServices(text string) ([]models.DetectedEvent, error) {
	ctx := context.Background()
	client, _ := genai.NewClient(ctx, option.WithAPIKey(config.AppConfig.GeminiKey))
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")

	prompt := fmt.Sprintf(`
		Extract all deadlines, exams, and recurring classes from the text.
		Return ONLY a JSON array. No conversational text.
		Use this format: [{"title": "Name", "start_time": "ISO8601", "confidence": 0.9, "source_text": "text from doc"}]
		
		TEXT: %s`, text)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	rawResponse := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// 2. Parse that string into our Go slice
	var events []models.DetectedEvent
	err = json.Unmarshal([]byte(rawResponse), &events)

	return events, err
}

func DetectEvents(text string) ([]models.DetectedEvent, error) {
	ctx := context.Background()

	client, _ := genai.NewClient(ctx, option.WithAPIKey(config.AppConfig.GeminiKey))
	defer client.Close()
	model := client.GenerativeModel("gemini-2.5-flash")

	prompt := fmt.Sprintf(`
        Extract all deadlines, exams, and recurring tasks from the text.
        Return ONLY a JSON array. 
        Use this format: [{"title": "Name", "start_time": "ISO8601", "confidence": 0.9, "source_text": "context snippet"}]
        
        TEXT: %s`, text)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	// AI often wraps JSON in code blocks, we need to clean it
	rawResponse := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	cleanedJSON := cleanAIJSON(rawResponse)

	var events []models.DetectedEvent
	err = json.Unmarshal([]byte(cleanedJSON), &events)

	return events, err
}

// Helper to remove AI markdown formatting
func cleanAIJSON(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "```json")
	input = strings.TrimPrefix(input, "```")
	input = strings.TrimSuffix(input, "```")
	return strings.TrimSpace(input)
}
