package models

import (
	"time"

	"github.com/google/uuid"
)

type DetectedEvent struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;index"`
	DocumentID uuid.UUID `gorm:"type:uuid;index"`
	Title      string
	StartTime  *time.Time
	EndTime    *time.Time
	Location   *string
	Confidence float64
	SourceText string
	DetectedAt time.Time
}
type Event struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Title     string
	StartTime time.Time
	EndTime   *time.Time
	Location  *string
	SourceID  uuid.UUID
	CreatedAt time.Time
}
