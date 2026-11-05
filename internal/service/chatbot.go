package service

import (
	"backend-ai-supernote/internal/constant"
	"backend-ai-supernote/internal/dto"
	entity "backend-ai-supernote/internal/entity/chatbot"
	repoNote "backend-ai-supernote/internal/repository"
	repository "backend-ai-supernote/internal/repository/chatbot"
	"backend-ai-supernote/pkg/chatbot"
	"backend-ai-supernote/pkg/embedding"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatbotService interface {
	CreateSession(ctx context.Context) (*dto.CreateSessionResponse, error)
	GetAllSessions(ctx context.Context) ([]*dto.GetAllSessionsResponse, error)
	GetChatHistory(ctx context.Context, sessionId uuid.UUID) ([]*dto.GetChatHistoryResponse, error)
	SendChat(ctx context.Context, req *dto.SendChatRequest) (*dto.SendChatResponse, error)
	DeleteSession(ctx context.Context, req *dto.DeleteSessionRequest) error
}

type chatbotService struct {
	db                 *pgxpool.Pool
	chatSessionRepo    repository.IChatSessionRepository
	chatMessageRepo    repository.IChatMessageRepository
	chatMessageRawRepo repository.IChatMessageRawRepository
	noteEmbeddingRepo  repoNote.INoteEmbeddingRepository
}

// DeleteSession implements [IChatbotService].
func (c *chatbotService) DeleteSession(ctx context.Context, req *dto.DeleteSessionRequest) error {
	_, err := c.chatSessionRepo.GetById(ctx, req.ChatSessionId)
	if err != nil {
		return err
	}
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	chatSessionRepoTx := c.chatSessionRepo.UsingTx(ctx, tx)
	chatMessageRepoTx := c.chatMessageRepo.UsingTx(ctx, tx)
	chatMessageRawRepoTx := c.chatMessageRawRepo.UsingTx(ctx, tx)

	err = chatSessionRepoTx.Delete(ctx, req.ChatSessionId)
	if err != nil {
		return err
	}

	err = chatMessageRepoTx.DeleteByChatSessionId(ctx, req.ChatSessionId)
	if err != nil {
		return err
	}

	err = chatMessageRawRepoTx.DeleteByChatSessionId(ctx, req.ChatSessionId)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
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
	noteEmbeddingRepoTx := c.noteEmbeddingRepo.UsingTx(ctx, tx)

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

	embeddingRes, err := embedding.GetGeminiEmbedding(
		os.Getenv("GEMINI_API_KEY"),
		req.Chat,
		"RETRIEVAL_QUERY",
	)
	if err != nil {
		return nil, err
	}

	decidedeUseRAGChatHistories := make([]*chatbot.ChatHistory, 0)
	for i, rawChat := range existingRawChats {
		if i == 0 {
			decidedeUseRAGChatHistories = append(decidedeUseRAGChatHistories, &chatbot.ChatHistory{
				Chat: constant.DecidedUseRAGMessageRawInitialUserPromptV1,
				Role: constant.ChatMessageRoleUser,
			})
			continue
		} else if i == 1 {
			decidedeUseRAGChatHistories = append(decidedeUseRAGChatHistories, &chatbot.ChatHistory{
				Chat: constant.DecidedUseRAGMessageRawInitialModelPromptV1,
				Role: constant.ChatMessageRoleUser,
			})
			continue
		}

		decidedeUseRAGChatHistories = append(decidedeUseRAGChatHistories, &chatbot.ChatHistory{
			Chat: rawChat.Chat,
			Role: rawChat.Role,
		})
	}
	useRag, err := chatbot.DecideToUseRAGE(
		ctx,
		os.Getenv("GEMINI_API_KEY"),
		decidedeUseRAGChatHistories,
	)
	if err != nil {
		return nil, err
	}
	strBuilder := strings.Builder{}

	if useRag {
		noteEmbeddings, err := noteEmbeddingRepoTx.SearchSimilarity(ctx, embeddingRes.Embedding.Values)
		if err != nil {
			return nil, err
		}

		for i, noteEmbedding := range noteEmbeddings {
			strBuilder.WriteString(fmt.Sprintf("Reference: &d\n", i+1))
			strBuilder.WriteString(noteEmbedding.Document)
			strBuilder.WriteString("\n\n")
		}
	}

	strBuilder.WriteString("User next question: ")
	strBuilder.WriteString(req.Chat)
	strBuilder.WriteString("\n\n")
	strBuilder.WriteString("Your answer? ")

	chatMessageRaw := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          strBuilder.String(),
		Role:          constant.ChatMessageRoleUser,
		ChatSessionId: req.ChatSessionId,
		CreatedAt:     now,
	}

	existingRawChats = append(existingRawChats, &chatMessageRaw)

	geminiReq := make([]*chatbot.ChatHistory, 0)
	for _, existingRawChat := range existingRawChats {
		geminiReq = append(geminiReq, &chatbot.ChatHistory{
			Chat: existingRawChat.Chat,
			Role: existingRawChat.Role,
		})
	}

	reply, err := chatbot.GetGeminiResponse(
		ctx,
		os.Getenv("GEMINI_API_KEY"),
		geminiReq,
	)
	if err != nil {
		return nil, err
	}

	chatMessageModel := entity.ChatMessage{
		Id:            uuid.New(),
		Chat:          reply,
		Role:          constant.ChatMessageRoleModel,
		ChatSessionId: req.ChatSessionId,
		CreatedAt:     now.Add(1 * time.Millisecond),
	}

	chatMessageRawModel := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          reply,
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
	noteEmbeddingRepo repoNote.INoteEmbeddingRepository,
) IChatbotService {
	return &chatbotService{
		db:                 db,
		chatSessionRepo:    chatSessionRepo,
		chatMessageRepo:    chatMessageRepo,
		chatMessageRawRepo: chatMessageRawRepo,
		noteEmbeddingRepo:  noteEmbeddingRepo,
	}
}
