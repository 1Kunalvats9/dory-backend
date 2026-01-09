package models

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;index"`
	Filename   string    `gorm:"size:255;not null"`
	FileURL    string    `gorm:"type:text"`
	PublicID   string
	FileType   string    `gorm:"size:50;default:'pdf'"`
	Content    string    `gorm:"type:text"`
	Status     string    `gorm:"size:20;default:'processing'"`
	UploadedAt time.Time `gorm:"autoCreateTime"`
}
