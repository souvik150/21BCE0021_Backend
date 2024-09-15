package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID           uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	FileName     string         `gorm:"not null;index"`
	OwnerID      uuid.UUID      `gorm:"not null"`
	Owner        User           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Size         int64          `gorm:"not null"`
	FileType     string         `gorm:"index"`
	CreatedAt    time.Time      `gorm:"index"`
	UpdatedAt    time.Time
	AccessedAt   time.Time
	DeletedStatus bool
	DeletedAt    time.Time `gorm:"index"`
}
