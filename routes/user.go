package routes

import (
	"att_service/controllers"
	"att_service/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App) {
	// Группа маршрутов для пользователей, требующая авторизации
	userGroup := app.Group("/user", middlewares.AuthMiddleware)
	// Редактирование профиля – страница student_page.html
	userGroup.Get("/profile", controllers.GetUserProfile)
	userGroup.Post("/maindata", controllers.SaveMainPageData)
	userGroup.Post("/change/photo", controllers.ChangeUserPhoto)
	// Загрузка документов – страница student_doc.html
	userGroup.Get("/document", controllers.GetUserDocuments)
	userGroup.Post("/documents/send", controllers.UploadUserDocuments)
	userGroup.Get("/data/correct", controllers.CheckUserDataCorrectness)
	userGroup.Post("/data/aprove", controllers.ApproveUserBySelf)
	userGroup.Get("/decline", controllers.GetDeclineReasons)
	userGroup.Get("/application", controllers.GetUserApplicationPage)
	userGroup.Get("/create-application", controllers.GetUserCreateApplicationPage)
	userGroup.Post("/create-application", controllers.SaveUserApplication)
}
