package repository

import (
	entity "backend-ai-supernote/internal/entity/chatbot"
	"backend-ai-supernote/pkg/database"
	"context"

	"github.com/google/uuid"
)

type IChatMessageRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRepository
	Create(ctx context.Context, chatMessage *entity.ChatMessage) error
	GetByChatSessionId(ctx context.Context, chatSessionId uuid.UUID) ([]*entity.ChatMessage, error)
}

type chatMessageRepository struct {
	db database.DatabaseQueryer
}

// GetByChatSessionId implements [IChatMessageRepository].
func (c *chatMessageRepository) GetByChatSessionId(ctx context.Context, chatSessionId uuid.UUID) ([]*entity.ChatMessage, error) {
	rows, err := c.db.Query(
		ctx,
		`SELECT
		id,
		role,
		chat,
		chat_session_id,
		created_at,
		updated_at,
		deleted_at,
		is_deleted
		FROM chat_message
		WHERE chat_session_id = $1
		AND
		is_deleted = false
		ORDER BY created_at ASC`,
		chatSessionId,
	)
	if err != nil {
		return nil, err
	}

	res := make([]*entity.ChatMessage, 0)
	for rows.Next() {
		var chatMessage entity.ChatMessage

		err = rows.Scan(
			&chatMessage.Id,
			&chatMessage.Role,
			&chatMessage.Chat,
			&chatMessage.ChatSessionId,
			&chatMessage.CreatedAt,
			&chatMessage.UpdatedAt,
			&chatMessage.DeletedAt,
			&chatMessage.IsDeleted,
		)

		if err != nil {
			return nil, err
		}

		res = append(res, &chatMessage)
	}

	return res, nil
}

// Create implements [IChatSessionRepository].
func (c *chatMessageRepository) Create(ctx context.Context, chatMessage *entity.ChatMessage) error {
	_, err := c.db.Exec(
		ctx,
		`INSERT INTO chat_message (
		id,
		role,
		chat,
		chat_session_id,
		created_at,
		updated_at,
		deleted_at,
		is_deleted)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		chatMessage.Id,
		chatMessage.Role,
		chatMessage.Chat,
		chatMessage.ChatSessionId,
		chatMessage.CreatedAt,
		chatMessage.UpdatedAt,
		chatMessage.DeletedAt,
		chatMessage.IsDeleted,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *chatMessageRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRepository {
	return &chatMessageRepository{
		db: tx,
	}
}

func NewChatMessageRepository(db database.DatabaseQueryer) IChatMessageRepository {
	return &chatMessageRepository{
		db: db,
	}
}
