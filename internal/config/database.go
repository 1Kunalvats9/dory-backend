package config

import (
	"dory-backend/internal/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(postgres.Open(AppConfig.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("error connecting to database", err)
	}
	err = database.AutoMigrate(&models.User{}, &models.Document{})
	if err != nil {
		log.Fatal("Migration failed: ", err)
	}
	log.Println("Database connected succesfully")

	DB = database
}
