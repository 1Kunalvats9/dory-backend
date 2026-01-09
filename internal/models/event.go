package models

import (
	"time"

	"github.com/google/uuid"
)

type DetectedEvent struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID  `gorm:"type:uuid;index"`
	DocumentID uuid.UUID  `gorm:"type:uuid;index"`
	Title      string     `gorm:"not null"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Recurrence string     `json:"recurrence"`
	Confidence float32    `json:"confidence"`
	SourceText string     `gorm:"type:text" json:"source_text"`
	CreatedAt  time.Time
}
