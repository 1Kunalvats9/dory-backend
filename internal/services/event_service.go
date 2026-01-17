package services

import (
	"context"
	"dory-backend/internal/config"
	"dory-backend/internal/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type inferredEventRaw struct {
	Title      string  `json:"title"`
	StartTime  *string `json:"start_time"`
	EndTime    *string `json:"end_time"`
	Location   *string `json:"location"`
	Confidence float64 `json:"confidence"`
	SourceText string  `json:"source_text"`
}

type InferredEvent struct {
	Title      string
	StartTime  *time.Time
	EndTime    *time.Time
	Location   *string
	Confidence float64
	SourceText string
}

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

func DetectEvents(text string) ([]InferredEvent, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(config.AppConfig.GeminiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")

	prompt := fmt.Sprintf(`
You are an event extraction engine.

Rules:
- Extract ONLY future or upcoming events
- Ignore past events, hypotheticals, examples
- Return ONLY valid JSON
- No markdown
- No explanations

Schema:
[
  {
    "title": "string",
    "start_time": "ISO8601 or null",
    "end_time": "ISO8601 or null",
    "location": "string or null",
    "confidence": number (0 to 1),
    "source_text": "exact snippet"
  }
]

If no events exist, return: []

TEXT:
%s
`, text)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 ||
		len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty LLM response")
	}

	raw := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	clean := cleanAIJSON(raw)

	var rawEvents []inferredEventRaw
	if err := json.Unmarshal([]byte(clean), &rawEvents); err != nil {
		return nil, err
	}

	// Normalize
	var results []InferredEvent
	for _, e := range rawEvents {
		var start, end *time.Time

		if e.StartTime != nil {
			if t, err := time.Parse(time.RFC3339, *e.StartTime); err == nil {
				start = &t
			}
		}

		if e.EndTime != nil {
			if t, err := time.Parse(time.RFC3339, *e.EndTime); err == nil {
				end = &t
			}
		}

		results = append(results, InferredEvent{
			Title:      e.Title,
			StartTime:  start,
			EndTime:    end,
			Location:   e.Location,
			Confidence: e.Confidence,
			SourceText: e.SourceText,
		})
	}

	return results, nil
}

func cleanAIJSON(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "```json")
	input = strings.TrimPrefix(input, "```")
	input = strings.TrimSuffix(input, "```")
	return strings.TrimSpace(input)
}
