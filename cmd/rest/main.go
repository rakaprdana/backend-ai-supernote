package main

import (
	"backend-ai-supernote/internal/controller"
	"backend-ai-supernote/internal/pkg/serverutils"
	"backend-ai-supernote/internal/repository"
	"backend-ai-supernote/internal/service"
	"backend-ai-supernote/pkg/database"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
	})

	app.Use(serverutils.ErrorHandlerMiddleware())

	db := database.ConnectDB(os.Getenv("DB_CONNECTION_STRING"))

	notebookRepository := repository.NewNotebookRepository(db)

	notebookService := service.NewNotebookService(notebookRepository)

	notebookController := controller.NewNotebookController(notebookService)

	api := app.Group("/api")
	notebookController.RegisterRoutes(api)

	log.Fatal(app.Listen(":3000"))
}
