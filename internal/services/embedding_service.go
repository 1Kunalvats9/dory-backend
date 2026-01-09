package services

import (
	"bytes"
	"dory-backend/internal/config"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func EmbedText(text string) ([]float32, error) {
	apiURL := "https://router.huggingface.co/hf-inference/models/intfloat/multilingual-e5-large/pipeline/feature-extraction"

	payload, _ := json.Marshal(map[string]interface{}{
		"inputs": text,
		"options": map[string]bool{
			"wait_for_model": true,
		},
	})

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+config.AppConfig.HuggingFaceToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HF API error: status %d", resp.StatusCode)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	switch v := result.(type) {
	case []interface{}:
		if len(v) > 0 {
			if first, ok := v[0].([]interface{}); ok {
				return convertToFloat32(first), nil
			}
			return convertToFloat32(v), nil
		}
	}

	return nil, errors.New("unexpected response format from HF")
}

func convertToFloat32(input []interface{}) []float32 {
	output := make([]float32, len(input))
	for i, val := range input {
		output[i] = float32(val.(float64))
	}
	return output
}
