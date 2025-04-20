package routes

import (
	"att_service/controllers"
	"att_service/middlewares"
	"att_service/services"

	"github.com/gofiber/websocket/v2"

	"github.com/gofiber/fiber/v2"
)

func RegisterPagesRoutes(app *fiber.App) {
	// Стартовая страница – логин (index.html)
	app.Get("/", controllers.LoginPage)
	// Страница регистрации
	app.Post("/login", controllers.Login)
	app.Post("/refresh", controllers.Refresh)
	app.Get("/ws", middlewares.AuthMiddleware, websocket.New(services.WebSocketHandler))
}
