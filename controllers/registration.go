package controllers

import (
	"att_service/database"
	"att_service/models"
	"att_service/services"
	"net/mail"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RegisterPage возвращает страницу регистрации (pages/registration.html)
func RegisterPage(c *fiber.Ctx) error {
	return c.SendFile("./views/pages/registration.html")
}

type RegistrationRequest struct {
	Email string `json:"email"`
}

func Register(c *fiber.Ctx) error {
	var req RegistrationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный формат данных"})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email обязателен"})
	}
	// Проверка валидности email
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный email адрес"})
	}

	// Проверка существования пользователя с таким email
	var existingUser models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Пользователь с таким email уже существует"})
	}

	if err := database.DB.Where("email = ?", req.Email).Delete(&models.EmailVerification{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка удаления старой записи подтверждения"})
	}

	// Генерация ссылки подтверждения
	link, err := services.GenerateVerificationLink(req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось сгенерировать ссылку"})
	}
	expiresAt := time.Now().Add(24 * time.Hour)

	verification := models.EmailVerification{
		Email:     req.Email,
		Link:      link,
		ExpiresAt: expiresAt,
	}
	if err := database.DB.Create(&verification).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка сохранения ссылки в базе данных"})
	}

	// Отправляем письмо с ссылкой подтверждения
	if err := services.SendVerificationEmail(req.Email, link); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось отправить email"})
	}

	return c.JSON(fiber.Map{"success": true})
}
