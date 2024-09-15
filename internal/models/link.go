package models

import (
	"time"

	"github.com/google/uuid"
)

type SharedLink struct {
	FileID     uuid.UUID `json:"file_id"`
	FileName   string    `json:"file_name"`
	ShareToken string    `json:"share_token" gorm:"primaryKey"`
	ExpiresAt  time.Time `json:"expires_at"`
}
