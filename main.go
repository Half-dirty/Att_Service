package main

import (
	"att_service/config"
	"att_service/controllers"
	"att_service/database"
	"att_service/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors" // Используем CORS middleware для Fiber
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
)

func main() {
	// Создаём HTML-шаблонизатор, указывая директорию и расширение шаблонов
	engine := html.New("C:/golang/att_service/views", ".html")
	engine.AddFunc("add1", func(i int) int { return i + 1 })
	engine.AddFunc("slice", func(args ...int) []int { return args })

	// Инициализируем Fiber с подключённым движком
	app := fiber.New(fiber.Config{
		Views:     engine,
		BodyLimit: 20 * 1024 * 1024,
	})

	// Настроим CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",                            // Разрешаем все источники для разработки (на проде поменяйте на конкретные домены)
		AllowMethods: "GET,POST,HEAD,PUT,DELETE",     // Разрешаем эти методы
		AllowHeaders: "Origin, Content-Type, Accept", // Разрешаем эти заголовки
	}))

	// Создаем store для сессий
	store := session.New()

	// Делаем его доступным во всех контроллерах
	controllers.SessionStore = store

	// Статические файлы
	app.Static("/", "./views/")
	app.Static("/style", "./views/style")
	app.Static("/scripts", "./views/scripts")
	app.Static("/pictures", "./views/pictures")
	app.Static("/uploads", "./uploads")

	// Регистрируем маршруты
	routes.RegisterPagesRoutes(app)
	routes.RegisterRegistrationRoutes(app)
	routes.RegisterUserRoutes(app)
	routes.RegisterAdminRoutes(app)

	// Подключаемся к базе данных
	if err := database.Connect(); err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// Запускаем сервер
	log.Fatal(app.Listen(config.ServerPort))
}
