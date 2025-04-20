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
	Quorum          int

	Examiners []User `gorm:"many2many:exam_examiners;"`
	Students  []User `gorm:"many2many:exam_students;"`
}
