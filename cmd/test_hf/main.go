package main

import (
	"dory-backend/internal/config"
	"dory-backend/internal/services"
	"fmt"
	"log"
)

func main() {
	config.LoadConfig()

	// 2. Define a test string
	testText := "Dory is a student assistant app."

	fmt.Println("--- Testing Hugging Face Embedding ---")
	fmt.Printf("Input: %s\n", testText)

	vectors, err := services.EmbedText(testText)
	if err != nil {
		log.Fatalf("Test Failed: %v", err)
	}

	fmt.Printf("Success! Received vector of length: %d\n", len(vectors))

	if len(vectors) > 0 {
		fmt.Printf("First 5 numbers: %v\n", vectors[:5])
	}

	if len(vectors) == 384 {
		fmt.Println("Result: Dimension count is correct (384).")
	} else {
		fmt.Printf("Result: Unexpected dimension count: %d\n", len(vectors))
	}
}
