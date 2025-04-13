package routes

import (
	"att_service/controllers"

	"github.com/gofiber/fiber/v2"
)

func RegisterPagesRoutes(app *fiber.App) {
	// Стартовая страница – логин (index.html)
	app.Get("/", controllers.LoginPage)
	// Страница регистрации
	app.Post("/login", controllers.Login)
}
