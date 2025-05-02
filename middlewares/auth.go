// package middlewares

// import (
// 	"att_service/config"
// 	"att_service/database"
// 	"att_service/models"
// 	"log"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/golang-jwt/jwt/v5"
// )

// // AuthMiddleware проверяет наличие и валидность access_token.
// func AuthMiddleware(c *fiber.Ctx) error {
// 	tokenString := c.Cookies("access_token")
// 	if tokenString == "" {
// 		log.Println("Отсутствует токен в куках")
// 		return c.Redirect("/?error=no_token")
// 	}

// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		return []byte(config.JWTSecret), nil
// 	})

// 	if err != nil {
// 		log.Println("Ошибка разбора JWT:", err)
// 		return c.Redirect("/?error=parse_error")
// 	}

// 	if !token.Valid {
// 		log.Println("Токен невалиден")
// 		return c.Redirect("/?error=invalid_token")
// 	}

// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok {
// 		log.Println("Ошибка извлечения claims из токена")
// 		return c.Redirect("/?error=claims_error")
// 	}

// 	userID, ok := claims["userID"].(float64)
// 	if !ok {
// 		log.Println("Ошибка типа userID в claims")
// 		return c.Redirect("/?error=invalid_userID")
// 	}

// 	c.Locals("userID", uint(userID))
// 	c.Locals("userEmail", claims["user"].(string))
// 	c.Locals("userRole", claims["role"].(string))

// 	return c.Next()
// }

// // AdminOnlyMiddleware разрешает доступ только администраторам.
// func AdminOnlyMiddleware(c *fiber.Ctx) error {
// 	role, ok := c.Locals("userRole").(string)
// 	if !ok || role != "admin" {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ только для администратора"})
// 	}
// 	return c.Next()
// }

// // CheckUserApprovalMiddleware проверяет, одобрен ли пользователь.
// // middlewares/auth.go
// func CheckUserApprovalMiddleware(c *fiber.Ctx) error {
// 	userIDVal := c.Locals("userID")
// 	if userIDVal == nil {
// 		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 			"error": "Необходима авторизация",
// 		})
// 	}

// 	var user models.User
// 	if err := database.DB.First(&user, userIDVal).Error; err != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 			"error": "Пользователь не найден",
// 		})
// 	}

// 	if user.Status != "approved" {
// 		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
// 			"error":  "Ваш аккаунт не одобрен администратором. Функция недоступна.",
// 			"status": user.Status,
// 		})
// 	}

//		return c.Next()
//	}
package middlewares

import (
	"att_service/config"
	"att_service/database"
	"att_service/models"
	"att_service/services"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// UniversalAuthMiddleware — единый middleware: проверяет access_token, обновляет при необходимости, проверяет роль и статус.
func UniversalAuthMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := validateAccessToken(c)
		if err != nil {
			user, err = tryRefreshToken(c)
			if err != nil {
				log.Println("[auth] access_token и refresh_token недействительны")
				clearAuthCookies(c)
				return c.Redirect("/")
			}
		}
		// Проверка на допустимые роли
		if len(allowedRoles) > 0 {
			allowed := false
			for _, r := range allowedRoles {
				if user.Role == r {
					allowed = true
					break
				}
			}
			if !allowed {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Недостаточно прав доступа",
				})
			}
		}

		// Установка данных в Locals
		c.Locals("userID", user.ID)
		c.Locals("userEmail", user.Email)
		c.Locals("userRole", user.Role)
		return c.Next()
	}
}

// validateAccessToken проверяет и возвращает пользователя по access_token.
func validateAccessToken(c *fiber.Ctx) (*models.User, error) {
	tokenStr := c.Cookies("access_token")
	if tokenStr == "" {
		return nil, fiber.ErrUnauthorized
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fiber.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}

	userID, ok := claims["userID"].(float64)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}

	var user models.User
	if err := database.DB.First(&user, uint(userID)).Error; err != nil {
		return nil, fiber.ErrUnauthorized
	}

	return &user, nil
}

// tryRefreshToken обновляет access_token через refresh_token из куки и возвращает пользователя.
func tryRefreshToken(c *fiber.Ctx) (*models.User, error) {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return nil, fiber.ErrUnauthorized
	}

	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fiber.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}

	email, ok := claims["user"].(string)
	if !ok {
		return nil, fiber.ErrUnauthorized
	}

	var user models.User
	if err := database.DB.Where("email = ? AND refresh_token = ?", email, refreshToken).First(&user).Error; err != nil {
		return nil, fiber.ErrUnauthorized
	}

	// Генерация нового access_token
	newAccessToken, _, err := services.GenerateTokens(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fiber.ErrUnauthorized
	}

	// Установка access_token в куку
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		HTTPOnly: false,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Now().Add(15 * time.Minute),
	})

	return &user, nil
}

// clearAuthCookies очищает куки авторизации (например, при недействительных токенах)
func clearAuthCookies(c *fiber.Ctx) {
	c.ClearCookie("access_token")
	c.ClearCookie("refresh_token")
}
