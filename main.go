package main

import (
	"att_service/config"
	"att_service/database"
	"att_service/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	_ "github.com/gofiber/template/html/v2"
)

func main() {
	// Создаём HTML-шаблонизатор, указывая директорию и расширение шаблонов
	engine := html.New("C:/golang/att_service/views", ".html")

	// Инициализируем Fiber с подключённым движком
	app := fiber.New(fiber.Config{
		Views:     engine,
		BodyLimit: 20 * 1024 * 1024,
	})

	app.Static("/", "./views/")
	app.Static("/scripts", "./views/scripts")
	app.Static("/style", "./views/style")
	app.Static("/pictures", "./views/pictures")
	app.Static("/uploads", "./uploads")

	// Регистрируем маршруты, которые работают только с предоставленными страницами
	routes.RegisterPagesRoutes(app)
	routes.RegisterRegistrationRoutes(app)
	routes.RegisterUserRoutes(app)
	routes.RegisterAdminRoutes(app)

	// Подключаемся к базе данных
	if err := database.Connect(); err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	log.Fatal(app.Listen(config.ServerPort))
}
