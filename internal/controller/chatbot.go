package controller

import (
	"backend-ai-supernote/internal/pkg/serverutils"
	"backend-ai-supernote/internal/service"

	"github.com/gofiber/fiber/v3"
)

type IChatbotController interface {
	RegisterRoutes(r fiber.Router)
	CreateSession(ctx fiber.Ctx) error
}

type chatbotController struct {
	chatbotService service.IChatbotService
}

// Create implements [INoteController].
func (c *chatbotController) CreateSession(ctx fiber.Ctx) error {

	res, err := c.chatbotService.CreateSession(ctx.Context())
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success create chat session", res))
}

// RegisterRoutes implements [INoteController].
func (c *chatbotController) RegisterRoutes(r fiber.Router) {
	h := r.Group("/chatbot/v1")
	h.Post("create-session", c.CreateSession)

}

func NewChatbotController(chatbotService service.IChatbotService) IChatbotController {
	return &chatbotController{chatbotService: chatbotService}
}
