package controllers

import (
	"att_service/config"
	"att_service/database"
	"att_service/models"
	"att_service/services"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Index перенаправляет на страницу входа.
func Index(c *fiber.Ctx) error {
	return c.Redirect("/login")
}

// LoginPage возвращает логин-страницу (pages/index.html)
func LoginPage(c *fiber.Ctx) error {
	return c.SendFile("./views/pages/index.html")
}

// Login обрабатывает форму входа и устанавливает access-токен в куки.
func Login(c *fiber.Ctx) error {
	// Вспомогательная структура для чтения JSON
	type LoginInput struct {
		Email string `json:"email"`
		Pass  string `json:"pass"`
	}
	var input LoginInput

	// Чтение и проверка тела запроса
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalidFormat"})
	}
	if input.Email == "" || input.Pass == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "emptyFields"})
	}

	// Поиск пользователя
	var existingUser models.User
	if err := database.DB.Where("email = ?", input.Email).First(&existingUser).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "userNone"})
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(input.Pass)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "passwordNone"})
	}

	// Генерация токенов
	accessToken, refreshToken, err := services.GenerateTokens(existingUser.ID, existingUser.Email, existingUser.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "tokenGeneration"})
	}

	// Сохранение refresh-токена
	if err := services.SaveRefreshToken(existingUser.Email, refreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "tokenSaveFail"})
	}

	// Установка access-токена в куки
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Now().Add(60 * time.Minute),
	})

	// Ответ клиенту
	if existingUser.Role == "admin" {
		return c.JSON(fiber.Map{"success": true, "link": "admin"})
	}
	return c.JSON(fiber.Map{"success": true, "link": "user/profile"})
}

// Logout очищает куки и перенаправляет на страницу входа.
func Logout(c *fiber.Ctx) error {
	c.ClearCookie("access_token")
	return c.Redirect("/")
}
func Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Отсутствует refresh-токен"})
	}

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неверный refresh-токен"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Ошибка токена"})
	}

	userEmail, ok := claims["user"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неверный токен"})
	}

	var user models.User
	if err := database.DB.Where("email = ? AND refresh_token = ?", userEmail, refreshToken).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Пользователь не найден"})
	}

	// Генерируем новые access и refresh токены
	newAccessToken, newRefreshToken, err := services.GenerateTokens(user.ID, user.Email, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка генерации токенов"})
	}

	// Сохраняем новый refresh токен в БД
	user.RefreshToken = newRefreshToken
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка сохранения токена"})
	}

	// Устанавливаем новые куки
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		HTTPOnly: false,
		Path:     "/",
		Secure:   false, // ❗ На HTTPS поставить true
		SameSite: "Strict",
		Expires:  time.Now().Add(15 * time.Minute), // Access токен короткий
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HTTPOnly: false,
		Path:     "/",
		Secure:   false,
		SameSite: "Strict",
		Expires:  time.Now().Add(7 * 24 * time.Hour), // Refresh токен на 7 дней
	})

	return c.JSON(fiber.Map{"success": true})
}

// Verify обрабатывает переход по ссылке подтверждения (GET /verify?token=...)
func Verify(c *fiber.Ctx) error {
	tokenString := c.Query("token")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return services.JwtSecret, nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusBadRequest).SendString("Неверный или просроченный токен")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("Неверный токен")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("Токен не содержит email")
	}

	expectedLink := "http://localhost:3000/verify?token=" + tokenString

	var verification models.EmailVerification
	if err := database.DB.Where("email = ? AND link = ?", email, expectedLink).First(&verification).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Запись подтверждения не найдена или уже использована")
	}
	if time.Now().After(verification.ExpiresAt) {
		return c.Status(fiber.StatusBadRequest).SendString("Срок действия ссылки истёк")
	}

	// Удаляем записи подтверждения для email
	if err := database.DB.
		Unscoped().
		Where("email = ?", email).
		Delete(&models.EmailVerification{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Не удалось очистить записи подтверждения")
	}

	// Теперь создаём пользователя
	user := models.User{
		Email:  email,
		Role:   "student",
		JestID: generateJestID(),
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка создания пользователя")
	}

	return services.Render(c, "", "confirm.html", fiber.Map{"email": email})
}

// генерация JestID
func generateJestID() string {
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	return fmt.Sprintf("06-10-%d", count+1)
}

// Confirm обрабатывает установку пароля (POST /confirm) и помечает пользователя как подтверждённого.
type ConfirmRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Confirm(c *fiber.Ctx) error {
	var req ConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Неверный формат данных")
	}
	log.Println(req.Email, req.Password)
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email и пароль обязательны")
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Пользователь не найден")
	}

	hashed, err := services.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка хеширования пароля")
	}
	user.Password = hashed
	user.Confirmed = true

	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка обновления пользователя")
	}

	return c.JSON(fiber.Map{"success": true})
}
