package models

import (
	"gorm.io/gorm"
)

type Token struct {
	gorm.Model
	User         string `gorm:"not null"` // maybe it will a user's email
	RefreshToken string `gorm:"not null"`
}
