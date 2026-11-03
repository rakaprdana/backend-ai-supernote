package service

import (
	"backend-ai-supernote/internal/dto"
	"backend-ai-supernote/internal/entity"
	"backend-ai-supernote/internal/repository"
	"backend-ai-supernote/pkg/embedding"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v3/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IConsumerService interface {
	Consume(ctx context.Context) error
}

type consumerService struct {
	notebookRepository repository.INotebookRepository
	noteRepository     repository.INoteRepository
	noteEmbeddingRepo  repository.INoteEmbeddingRepository
	pubSub             *gochannel.GoChannel
	topicName          string
	db                 *pgxpool.Pool
}

// Consume implements [IConsumerService].
func (c *consumerService) Consume(ctx context.Context) error {
	messages, err := c.pubSub.Subscribe(ctx, c.topicName)
	if err != nil {
		return err
	}

	go func() {
		for msg := range messages {
			c.processMessage(ctx, msg)
		}

	}()

	return nil
}

func NewConsumerService(
	notebookRepository repository.INotebookRepository,
	noteRepository repository.INoteRepository,
	noteEmbeddingRepo repository.INoteEmbeddingRepository,
	pubSub *gochannel.GoChannel,
	topicName string,
	db *pgxpool.Pool,
) IConsumerService {
	return &consumerService{
		notebookRepository: notebookRepository,
		noteRepository:     noteRepository,
		noteEmbeddingRepo:  noteEmbeddingRepo,
		pubSub:             pubSub,
		topicName:          topicName,
		db:                 db,
	}
}

func (c *consumerService) processMessage(
	ctx context.Context,
	msg *message.Message,
) {
	defer msg.Nack()
	defer func() {
		if e := recover(); e != nil {
			log.Error(e)
		}
	}()

	var payload dto.PublisherEmbedNoteMessage
	err := json.Unmarshal(msg.Payload, &payload)
	if err != nil {
		panic(err)
	}

	note, err := c.noteRepository.GetByID(ctx, payload.NoteId)
	if err != nil {
		panic(err)
	}

	notebook, err := c.notebookRepository.GetByID(ctx, note.NotebookId)
	if err != nil {
		panic(err)
	}

	noteUpdated := "-"
	if note.UpdatedAt != nil {
		noteUpdated = note.UpdatedAt.Format(time.RFC3339)
	}
	content := fmt.Sprintf(`
	Note Title : %s
	Notebook Title: %s

	%s

	Created At: %s
	Updated At: %s
	`,
		note.Title,
		notebook.Name,
		note.CreatedAt.Format(time.RFC3339),
		note.Content,
		noteUpdated,
	)

	result, err := embedding.GetGeminiEmbedding(
		os.Getenv("GEMINI_API_KEY"),
		content,
		"RETRIEVAL_DOCUMENT",
	)
	if err != nil {
		panic(err)
	}

	noteEmbedding := entity.NoteEmbedding{
		Id:             uuid.New(),
		Document:       content,
		EmbeddingValue: result.Embedding.Values,
		NoteId:         note.Id,
		CreatedAt:      time.Now(),
	}

	tx, err := c.db.Begin(ctx)
	if err != nil {
		panic(err)
	}

	defer tx.Rollback(ctx)

	noteEmbeddingRepository := c.noteEmbeddingRepo.UsingTx(ctx, tx)

	err = noteEmbeddingRepository.DeleteByNoteId(ctx, note.Id)
	if err != nil {
		panic(err)
	}

	err = noteEmbeddingRepository.Create(ctx, &noteEmbedding)
	if err != nil {
		panic(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		panic(err)
	}

	msg.Ack()
}
