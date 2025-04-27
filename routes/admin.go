package routes

import (
	"att_service/controllers"
	"att_service/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterAdminRoutes(app *fiber.App) {
	// Группа маршрутов для администратора (проверяется middleware)
	adminGroup := app.Group("/admin", middlewares.UniversalAuthMiddleware("admin"))
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
	adminGroup.Get("/exam/students", controllers.GetAdminExamApplications)
	adminGroup.Get("/student/application/:id", controllers.AdminShowStudentApplication)
	adminGroup.Post("/api/application/approve", controllers.ApproveApplication)
	adminGroup.Get("/exam/scheduled", controllers.ExamScheduledPage)
	adminGroup.Post("/api/exam/schedule", controllers.ScheduleExam)
	adminGroup.Post("/api/exam/set", controllers.AdminSetTargetExam)
	adminGroup.Get("/exam/show", controllers.AdminShowExam)
	adminGroup.Post("/api/exam/cancel", controllers.AdminCancelExam)
	adminGroup.Post("/api/application/decline", controllers.DeclineApplication)
	adminGroup.Post("/student/decline", controllers.AdminDeclineStudent)
	adminGroup.Get("/exam/view", controllers.AdminViewExam)

}
