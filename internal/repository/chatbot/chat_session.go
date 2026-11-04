package repository

import (
	entity "backend-ai-supernote/internal/entity/chatbot"
	"backend-ai-supernote/pkg/database"
	"context"
)

type IChatSessionRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatSessionRepository
	Create(ctx context.Context, chatSession *entity.ChatSession) error
}

type chatSessionRepository struct {
	db database.DatabaseQueryer
}

// Create implements [IChatSessionRepository].
func (c *chatSessionRepository) Create(ctx context.Context, chatSession *entity.ChatSession) error {
	_, err := c.db.Exec(
		ctx,
		`INSERT INTO chat_session (id, title, created_at, updated_at, deleted_at, is_deleted)
	VALUES ($1, $2, $3, $4, $5, $6)`,
		chatSession.Id,
		chatSession.Title,
		chatSession.CreatedAt,
		chatSession.UpdatedAt,
		chatSession.DeletedAt,
		chatSession.IsDeleted,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *chatSessionRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatSessionRepository {
	return &chatSessionRepository{
		db: tx,
	}
}

func NewChatSessionRepository(db database.DatabaseQueryer) IChatSessionRepository {
	return &chatSessionRepository{
		db: db,
	}
}
