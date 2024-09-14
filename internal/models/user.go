package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Email     string         `gorm:"unique;not null"`
	Password  string         `gorm:"not null"`
	Files     []File         `gorm:"foreignKey:OwnerID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
