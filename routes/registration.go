package routes

import (
	"att_service/controllers"

	"github.com/gofiber/fiber/v2"
)

func RegisterRegistrationRoutes(app *fiber.App) {
	// Форма регистрации
	app.Get("/registration", controllers.RegisterPage)
	// Обработка регистрации по email
	app.Post("/registration", controllers.Register)
	// Переход по ссылке подтверждения
	app.Get("/verify", controllers.Verify)
	// Обработка установки пароля и подтверждения email
	app.Post("/confirm", controllers.Confirm)
}
