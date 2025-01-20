package models

import (
	"github.com/google/uuid"

	"github.com/jkulzer/fib-server/sharedModels"

	"gorm.io/gorm"
)

type LoginInfo struct {
	gorm.Model
	ID         uint `gorm:"primaryKey;autoIncrement"`
	Token      uuid.UUID
	LobbyToken string
	Role       sharedModels.UserRole
}

var NullUuidString = "00000000-0000-0000-0000-000000000000"
