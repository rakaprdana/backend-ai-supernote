package cmd_rest

import (
	"backend-ai-supernote/internal/controller"
	"backend-ai-supernote/internal/pkg/serverutils"
	"backend-ai-supernote/internal/repository"
	chatbotRepo "backend-ai-supernote/internal/repository/chatbot"
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
	noteEmbeddingRepository := repository.NewNoteEmbeddingRepository(db)
	chatSessionRepository := chatbotRepo.NewChatSessionRepository(db)
	chatMessageRepository := chatbotRepo.NewChatMessageRepository(db)
	chatMessagRawRepository := chatbotRepo.NewChatMessageRawRepository(db)

	watermillLogger := watermill.NewStdLogger(false, false)
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermillLogger,
	)
	publisherService := service.NewPublisherService(
		pubSub,
		os.Getenv("EMBED_NOTE_CONTENT_TOPIC_NAME"),
	)
	consumerService := service.NewConsumerService(
		notebookRepository,
		noteRepository,
		noteEmbeddingRepository,
		pubSub,
		os.Getenv("EMBED_NOTE_CONTENT_TOPIC_NAME"),
		db,
	)

	notebookService := service.NewNotebookService(
		notebookRepository,
		noteRepository,
		noteEmbeddingRepository,
		publisherService,
		db,
	)

	noteService := service.NewNoteService(
		noteRepository,
		noteEmbeddingRepository,
		publisherService,
		db,
	)
	chatbotService := service.NewChatbotService(
		db,
		chatSessionRepository,
		chatMessageRepository,
		chatMessagRawRepository,
		noteEmbeddingRepository,
	)

	notebookController := controller.NewNotebookController(notebookService)
	noteController := controller.NewNoteController(noteService)
	chatbotController := controller.NewChatbotController(chatbotService)

	api := app.Group("/api")
	notebookController.RegisterRoutes(api)
	noteController.RegisterRoutes(api)
	chatbotController.RegisterRoutes(api)

	err := consumerService.Consume(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("Server running")
	log.Fatal(app.Listen(":3000"))
}
