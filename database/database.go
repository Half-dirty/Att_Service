package database

import (
	"fmt"
	"log"

	"att_service/config"
	"att_service/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		config.DBHost,
		config.DBUser,
		config.DBPassword,
		config.DBName,
		config.DBPort,
		config.DBSSLMode,
		config.DBTimeZone,
	)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// Автомиграция моделей
	if err := DB.AutoMigrate(
		&models.User{},
		&models.UserDocument{},
		&models.EmailVerification{},
		&models.Passport{},
		&models.EducationDocument{},
		&models.Application{},
		&models.Exam{},
		&models.ApplicationDecline{},
		&models.ExamExaminer{},
		&models.ExamStudent{},
	); err != nil {
		return err
	}
	log.Println("Соединение с базой данных установлено и миграции успешно выполнены")
	return nil
}

func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Println("Ошибка получения sqlDB:", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Fatalf("Ошибка закрытия БД: %v", err)
	}
	log.Println("Соединение с БД закрыто")
}
