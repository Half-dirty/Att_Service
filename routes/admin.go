package routes

import (
	"att_service/controllers"
	"att_service/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterAdminRoutes(app *fiber.App) {
	// Группа маршрутов для администратора (проверяется middleware)
	adminGroup := app.Group("/admin", middlewares.AuthMiddleware, middlewares.AdminOnlyMiddleware)
	// Главная страница админа – admin_page.html
	adminGroup.Get("/", controllers.AdminPage)
	adminGroup.Get("/user/list", controllers.AdminUserList)
	adminGroup.Get("/show/student/:id", controllers.AdminShowStudentProfile)
	adminGroup.Post("/select/application", controllers.AdminSelectApplicationUsers)
	adminGroup.Post("/search/all", controllers.AdminSearchAll)
	adminGroup.Get("/user/application", controllers.AdminUserApplication)
	adminGroup.Post("/select", controllers.AdminSelectUsersByRole)
	adminGroup.Get("/show/document/:id", controllers.AdminShowStudentDocuments)
	adminGroup.Post("/student/confirm/:id", controllers.AdminConfirmStudent)
	adminGroup.Post("/decline/:id", controllers.AdminDeclineStudent)
	adminGroup.Post("/delete/student/:id", controllers.AdminDeleteStudent)
	adminGroup.Post("/change_role", controllers.AdminChangeUserRole)
}
