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
	adminGroup.Post("/select/application", controllers.AdminSelectApplicationUsers)
	adminGroup.Post("/search/all", controllers.AdminSearchAll)
	adminGroup.Get("/user/application", controllers.AdminUserApplication)
	adminGroup.Post("/select", controllers.AdminSelectUsersByRole)
	adminGroup.Post("/change_role", controllers.AdminChangeUserRole)
	adminGroup.Get("/exam/list", controllers.GetPastExamsPage)
	adminGroup.Get("/exam/planning", controllers.ExamPlanningPage)
	adminGroup.Get("/exam/create", controllers.AdminCreateExamPage)
	adminGroup.Post("/exam/create", controllers.AdminCreateExam)
	adminGroup.Post("/api/student", controllers.AdminSetTargetStudent) // установка targetStudentID
	adminGroup.Get("/student/profile", controllers.AdminShowStudentProfile)
	adminGroup.Get("/student/documents", controllers.AdminShowStudentDocuments)
	adminGroup.Post("/student/confirm", controllers.AdminConfirmStudent)
	adminGroup.Post("/student/decline", controllers.AdminDeclineStudent)
	adminGroup.Post("/student/delete", controllers.AdminDeleteStudent)
}
