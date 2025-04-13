package middlewares

import (
	"att_service/config"
	"att_service/database"
	"att_service/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware проверяет наличие и валидность access_token.
func AuthMiddleware(c *fiber.Ctx) error {
	tokenString := c.Cookies("access_token")
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Необходимо авторизоваться"})
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Недействительный токен"})
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Ошибка токена"})
	}
	userID := uint(claims["userID"].(float64))
	userEmail := claims["user"].(string)
	userRole := claims["role"].(string)
	c.Locals("userID", userID)
	c.Locals("userEmail", userEmail)
	c.Locals("userRole", userRole)
	return c.Next()
}

// AdminOnlyMiddleware разрешает доступ только администраторам.
func AdminOnlyMiddleware(c *fiber.Ctx) error {
	role, ok := c.Locals("userRole").(string)
	if !ok || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ только для администратора"})
	}
	return c.Next()
}

// CheckUserApprovalMiddleware проверяет, одобрен ли пользователь.
// middlewares/auth.go
func CheckUserApprovalMiddleware(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Необходима авторизация",
		})
	}

	var user models.User
	if err := database.DB.First(&user, userIDVal).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Пользователь не найден",
		})
	}

	if user.Status != "approved" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":  "Ваш аккаунт не одобрен администратором. Функция недоступна.",
			"status": user.Status,
		})
	}

	return c.Next()
}
