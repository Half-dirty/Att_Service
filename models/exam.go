package models

import (
	"time"

	"gorm.io/gorm"
)

type Exam struct {
	gorm.Model
	Date            time.Time
	CommissionStart time.Time
	CommissionEnd   time.Time
	Status          string `gorm:"size:50;not null;default:'planned'"` // planned, scheduled, completed
	ChairmanID      uint
	SecretaryID     uint
	Quorum          int
	JestID          string `gorm:"unique;size:20;column:jest_id"`

	Examiners []User `gorm:"many2many:exam_examiners;"`
	Students  []User `gorm:"many2many:exam_students;"`
}

type ExamStudent struct {
	ID     uint `gorm:"primaryKey"`
	JestID string
	UserID uint
	ExamID uint // 🔥 обязательно добавляем
}

type ExamExaminer struct {
	ID     uint `gorm:"primaryKey"`
	JestID string
	UserID uint
	ExamID uint // 🔥 обязательно добавляем
}
