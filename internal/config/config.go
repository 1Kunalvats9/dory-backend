package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	JWTSecret         string
	GeminiKey         string
	GoogleWebClientID string
	GoogleIOSClientID string
	CloudinaryURL     string
	HuggingFaceToken  string
	QdrantHost        string
	QdrantKey         string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env found")
	}

	AppConfig = &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		GeminiKey:         os.Getenv("GEMINI_API_KEY"),
		GoogleWebClientID: os.Getenv("GOOGLE_WEB_CLIENT_ID"),
		GoogleIOSClientID: os.Getenv("GOOGLE_IOS_CLIENT_ID"),
		CloudinaryURL:     os.Getenv("CLOUDINARY_URL"),
		HuggingFaceToken:  os.Getenv("HUGGING_FACE_TOKEN"),
		QdrantHost:        os.Getenv("QDRANT_HOST"),
		QdrantKey:         os.Getenv("QDRANT_API_KEY"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
