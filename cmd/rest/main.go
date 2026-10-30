package cmd_rest

import (
	"backend-ai-supernote/internal/controller"
	"backend-ai-supernote/internal/pkg/serverutils"
	"backend-ai-supernote/internal/repository"
	"backend-ai-supernote/internal/service"
	"backend-ai-supernote/pkg/database"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
)

func Main() {
	godotenv.Load()
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
	})

	app.Use(cors.New())
	app.Use(serverutils.ErrorHandlerMiddleware())

	db := database.ConnectDB(os.Getenv("DB_CONNECTION_STRING"))

	notebookRepository := repository.NewNotebookRepository(db)
	noteRepository := repository.NewNoteRepository(db)

	watermillLogger := watermill.NewStdLogger(false, false)
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermillLogger,
	)
	publisherService := service.NewPublisherService(
		pubSub,
		"embed-note-content",
	)
	consumerService := service.NewConsumerService(
		pubSub,
		"embed-note-content",
	)

	notebookService := service.NewNotebookService(notebookRepository, noteRepository, db)
	noteService := service.NewNoteService(noteRepository, publisherService)

	notebookController := controller.NewNotebookController(notebookService)
	noteController := controller.NewNoteController(noteService)

	api := app.Group("/api")
	notebookController.RegisterRoutes(api)
	noteController.RegisterRoutes(api)
	err := consumerService.Consume(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("Server running")
	log.Fatal(app.Listen(":3000"))
}
