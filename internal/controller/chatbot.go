package controller

import (
	"backend-ai-supernote/internal/dto"
	"backend-ai-supernote/internal/pkg/serverutils"
	"backend-ai-supernote/internal/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type IChatbotController interface {
	RegisterRoutes(r fiber.Router)
	CreateSession(ctx fiber.Ctx) error
	GetAllSessions(ctx fiber.Ctx) error
	GetChatHistory(ctx fiber.Ctx) error
	SendChat(ctx fiber.Ctx) error
}

type chatbotController struct {
	chatbotService service.IChatbotService
}

// SendChat implements [IChatbotController].
func (c *chatbotController) SendChat(ctx fiber.Ctx) error {
	var req dto.SendChatRequest

	if err := ctx.Bind().Body(&req); err != nil {
		return err
	}

	if err := serverutils.ValidateRequest(req); err != nil {
		return err
	}

	res, err := c.chatbotService.SendChat(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success send chat", res))
}

// GetChatHistory implements [IChatbotController].
func (c *chatbotController) GetChatHistory(ctx fiber.Ctx) error {
	idStr := ctx.Query("chat_session_id")
	sessionId, _ := uuid.Parse(idStr)

	res, err := c.chatbotService.GetChatHistory(ctx.Context(), sessionId)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success get chat history", res))
}

// GetAllSessions implements [IChatbotController].
func (c *chatbotController) GetAllSessions(ctx fiber.Ctx) error {
	res, err := c.chatbotService.GetAllSessions(ctx.Context())
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success get all session", res))
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
	h.Get("sessions", c.GetAllSessions)
	h.Get("chat-history", c.GetChatHistory)
	h.Post("send-chat", c.SendChat)

}

func NewChatbotController(chatbotService service.IChatbotService) IChatbotController {
	return &chatbotController{chatbotService: chatbotService}
}
