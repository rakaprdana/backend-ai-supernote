package repository

import (
	entity "backend-ai-supernote/internal/entity/chatbot"
	"backend-ai-supernote/pkg/database"
	"context"

	"github.com/google/uuid"
)

type IChatMessageRawRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRawRepository
	Create(ctx context.Context, chatMessageRaw *entity.ChatMessageRaw) error
	GetByChatSessionId(ctx context.Context, chatSessionId uuid.UUID) ([]*entity.ChatMessageRaw, error)
}

type chatMessageRawRepository struct {
	db database.DatabaseQueryer
}

// GetByChatSessionId implements [IChatMessageRawRepository].
func (c *chatMessageRawRepository) GetByChatSessionId(ctx context.Context, chatSessionId uuid.UUID) ([]*entity.ChatMessageRaw, error) {
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
		FROM chat_message_raw
		WHERE chat_session_id = $1
		AND
		is_deleted = false
		ORDER BY created_at ASC`,
		chatSessionId,
	)
	if err != nil {
		return nil, err
	}

	res := make([]*entity.ChatMessageRaw, 0)
	for rows.Next() {
		var chatMessageRaw entity.ChatMessageRaw

		err = rows.Scan(
			&chatMessageRaw.Id,
			&chatMessageRaw.Role,
			&chatMessageRaw.Chat,
			&chatMessageRaw.ChatSessionId,
			&chatMessageRaw.CreatedAt,
			&chatMessageRaw.UpdatedAt,
			&chatMessageRaw.DeletedAt,
			&chatMessageRaw.IsDeleted,
		)

		if err != nil {
			return nil, err
		}

		res = append(res, &chatMessageRaw)
	}

	return res, nil
}

// Create implements [IChatSessionRepository].
func (c *chatMessageRawRepository) Create(ctx context.Context, chatMessageRaw *entity.ChatMessageRaw) error {
	_, err := c.db.Exec(
		ctx,
		`INSERT INTO chat_message_raw (
		id,
		role,
		chat,
		chat_session_id,
		created_at,
		updated_at,
		deleted_at,
		is_deleted)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		chatMessageRaw.Id,
		chatMessageRaw.Role,
		chatMessageRaw.Chat,
		chatMessageRaw.ChatSessionId,
		chatMessageRaw.CreatedAt,
		chatMessageRaw.UpdatedAt,
		chatMessageRaw.DeletedAt,
		chatMessageRaw.IsDeleted,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *chatMessageRawRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRawRepository {
	return &chatMessageRawRepository{
		db: tx,
	}
}

func NewChatMessageRawRepository(db database.DatabaseQueryer) IChatMessageRawRepository {
	return &chatMessageRawRepository{
		db: db,
	}
}
