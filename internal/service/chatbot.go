package service

import (
	"backend-ai-supernote/internal/constant"
	"backend-ai-supernote/internal/dto"
	entity "backend-ai-supernote/internal/entity/chatbot"
	repository "backend-ai-supernote/internal/repository/chatbot"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatbotService interface {
	CreateSession(ctx context.Context) (*dto.CreateSessionResponse, error)
	GetAllSessions(ctx context.Context) ([]*dto.GetAllSessionsResponse, error)
	GetChatHistory(ctx context.Context, sessionId uuid.UUID) ([]*dto.GetChatHistoryResponse, error)
	SendChat(ctx context.Context, req *dto.SendChatRequest) (*dto.SendChatResponse, error)
}

type chatbotService struct {
	db                 *pgxpool.Pool
	chatSessionRepo    repository.IChatSessionRepository
	chatMessageRepo    repository.IChatMessageRepository
	chatMessageRawRepo repository.IChatMessageRawRepository
}

// SendChat implements [IChatbotService].
func (c *chatbotService) SendChat(ctx context.Context, req *dto.SendChatRequest) (*dto.SendChatResponse, error) {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	chatSessionRepoTx := c.chatSessionRepo.UsingTx(ctx, tx)
	chatMessageRepoTx := c.chatMessageRepo.UsingTx(ctx, tx)
	chatMessageRawRepoTx := c.chatMessageRawRepo.UsingTx(ctx, tx)

	chatSession, err := chatSessionRepoTx.GetById(ctx, req.ChatSessionId)
	if err != nil {
		return nil, err
	}

	existingRawChats, err := c.chatMessageRawRepo.GetByChatSessionId(ctx, req.ChatSessionId)
	if err != nil {
		return nil, err
	}
	updateSessiontTitle := len(existingRawChats) == 2

	now := time.Now()

	chatMessage := entity.ChatMessage{
		Id:            uuid.New(),
		Chat:          req.Chat,
		Role:          constant.ChatMessageRoleUser,
		ChatSessionId: req.ChatSessionId,
		CreatedAt:     now,
	}

	chatMessageRaw := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          req.Chat,
		Role:          constant.ChatMessageRoleUser,
		ChatSessionId: req.ChatSessionId,
		CreatedAt:     now,
	}

	chatMessageModel := entity.ChatMessage{
		Id:            uuid.New(),
		Chat:          "This is automated dummy response",
		Role:          constant.ChatMessageRoleModel,
		ChatSessionId: req.ChatSessionId,
		CreatedAt:     now.Add(1 * time.Millisecond),
	}

	chatMessageRawModel := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          "This is automated dummy response",
		Role:          constant.ChatMessageRoleModel,
		ChatSessionId: req.ChatSessionId,
		CreatedAt:     now.Add(1 * time.Millisecond),
	}

	chatMessageRepoTx.Create(ctx, &chatMessage)
	chatMessageRepoTx.Create(ctx, &chatMessageModel)
	chatMessageRawRepoTx.Create(ctx, &chatMessageRaw)
	chatMessageRawRepoTx.Create(ctx, &chatMessageRawModel)

	if updateSessiontTitle {
		chatSession.Title = req.Chat
		chatSession.UpdatedAt = &now

		err = chatSessionRepoTx.Update(ctx, chatSession)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	res := &dto.SendChatResponse{
		ChatSessionId:    chatSession.Id,
		ChatSessionTitle: chatSession.Title,
		Sent: &dto.SendChatResponseChat{
			Id:        chatMessage.Id,
			Chat:      chatMessage.Chat,
			Role:      chatMessage.Role,
			CreatedAt: chatMessage.CreatedAt,
		},
		Reply: &dto.SendChatResponseChat{
			Id:        chatMessageModel.Id,
			Chat:      chatMessageModel.Chat,
			Role:      chatMessageModel.Role,
			CreatedAt: chatMessageModel.CreatedAt,
		},
	}
	return res, nil
}

// GetChatHistory implements [IChatbotService].
func (c *chatbotService) GetChatHistory(ctx context.Context, sessionId uuid.UUID) ([]*dto.GetChatHistoryResponse, error) {
	_, err := c.chatSessionRepo.GetById(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	chatMessages, err := c.chatMessageRepo.GetByChatSessionId(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	res := make([]*dto.GetChatHistoryResponse, 0)
	for _, chatMessage := range chatMessages {
		res = append(res, &dto.GetChatHistoryResponse{
			Id:        chatMessage.Id,
			Role:      chatMessage.Role,
			Chat:      chatMessage.Chat,
			CreatedAt: chatMessage.CreatedAt,
		})
	}

	return res, nil
}

// GetAllSessions implements [IChatbotService].
func (c *chatbotService) GetAllSessions(ctx context.Context) ([]*dto.GetAllSessionsResponse, error) {
	chatSessions, err := c.chatSessionRepo.GetAll(ctx)

	if err != nil {
		return nil, err
	}

	res := make([]*dto.GetAllSessionsResponse, 0)
	for _, chatSession := range chatSessions {
		res = append(res, &dto.GetAllSessionsResponse{
			Id:        chatSession.Id,
			Title:     chatSession.Title,
			CreatedAt: chatSession.CreatedAt,
			UpdatedAt: chatSession.UpdatedAt,
		})
	}

	return res, nil
}

func (c *chatbotService) CreateSession(ctx context.Context) (*dto.CreateSessionResponse, error) {
	now := time.Now()

	chatSession := entity.ChatSession{
		Id:        uuid.New(),
		Title:     "Unnamed session",
		CreatedAt: now,
	}

	chatMessage := entity.ChatMessage{
		Id:            uuid.New(),
		Chat:          "Hi, how can I help you ?",
		Role:          constant.ChatMessageRoleModel,
		ChatSessionId: chatSession.Id,
		CreatedAt:     now,
	}

	chatMessageRawUser := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          constant.ChatMessageRawInitialUserPromptV1,
		Role:          constant.ChatMessageRoleUser,
		ChatSessionId: chatSession.Id,
		CreatedAt:     now,
	}

	chatMessageRawModel := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          constant.ChatMessageRawInitialModelPromptV1,
		Role:          constant.ChatMessageRoleModel,
		ChatSessionId: chatSession.Id,
		CreatedAt:     now.Add(1 * time.Second),
	}

	tx, err := c.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	chatSessionRepoTx := c.chatSessionRepo.UsingTx(ctx, tx)
	chatMessageRepoTx := c.chatMessageRepo.UsingTx(ctx, tx)
	chatMessageRawRepoTx := c.chatMessageRawRepo.UsingTx(ctx, tx)

	if err := chatSessionRepoTx.Create(ctx, &chatSession); err != nil {
		return nil, err
	}
	if err := chatMessageRepoTx.Create(ctx, &chatMessage); err != nil {
		return nil, err
	}
	if err := chatMessageRawRepoTx.Create(ctx, &chatMessageRawUser); err != nil {
		return nil, err
	}
	if err := chatMessageRawRepoTx.Create(ctx, &chatMessageRawModel); err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	res := &dto.CreateSessionResponse{
		Id: chatSession.Id,
	}

	return res, nil
}

func NewChatbotService(
	db *pgxpool.Pool,
	chatSessionRepo repository.IChatSessionRepository,
	chatMessageRepo repository.IChatMessageRepository,
	chatMessageRawRepo repository.IChatMessageRawRepository,
) IChatbotService {
	return &chatbotService{
		db:                 db,
		chatSessionRepo:    chatSessionRepo,
		chatMessageRepo:    chatMessageRepo,
		chatMessageRawRepo: chatMessageRawRepo,
	}
}
