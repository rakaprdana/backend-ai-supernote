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
}

type chatbotService struct {
	db                 *pgxpool.Pool
	chatSessionRepo    repository.IChatSessionRepository
	chatMessageRepo    repository.IChatMessageRepository
	chatMessageRawRepo repository.IChatMessageRawRepository
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
