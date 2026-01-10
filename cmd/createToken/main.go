package main

import (
	"dory-backend/internal/config"
	"dory-backend/internal/services"
	"fmt"
	"log"
)

func main() {
	// 1. Load config to get your JWT_SECRET
	config.LoadConfig()

	// 2. Generate a token for a mock user
	// We use a fake UUID and email just for testing
	testUserID := "00000000-0000-0000-0000-000000000001"
	testEmail := "testuser@dory.com"

	token, err := services.GenerateJWTToken(testUserID, testEmail)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	fmt.Println("--- Dory Test Token ---")
	fmt.Println("Copy and use this in your curl command:")
	fmt.Println("")
	fmt.Printf("Bearer %s\n", token)
	fmt.Println("")
}
