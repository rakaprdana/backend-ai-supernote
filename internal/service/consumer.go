package service

import (
	"backend-ai-supernote/internal/dto"
	"backend-ai-supernote/internal/repository"
	"backend-ai-supernote/pkg/database/embedding"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v3/log"
)

type IConsumerService interface {
	Consume(ctx context.Context) error
}

type consumerService struct {
	noteRepository repository.INoteRepository
	pubSub         *gochannel.GoChannel
	topicName      string
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
	noteRepository repository.INoteRepository,
	pubSub *gochannel.GoChannel,
	topicName string,
) IConsumerService {
	return &consumerService{
		noteRepository: noteRepository,
		pubSub:         pubSub,
		topicName:      topicName,
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

	result, err := embedding.GetGeminiEmbedding(os.Getenv("GEMINI_API_KEY"), note.Content)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Embedding.Values)
	msg.Ack()
}
