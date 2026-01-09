package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string     `gorm:"uniqueIndex;not null"`
	Name         string     `gorm:"size:255"`
	GoogleID     string     `gorm:"uniqueIndex"`
	ProfilePhoto string     `gorm:"type:text"`
	Documents    []Document `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
