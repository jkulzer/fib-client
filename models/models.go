package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginInfo struct {
	gorm.Model
	ID         uint `gorm:"primaryKey;autoIncrement"`
	Token      uuid.UUID
	LobbyToken string
}
