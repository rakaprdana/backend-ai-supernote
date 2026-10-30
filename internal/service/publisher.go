package service

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type IPublisherService interface {
	Publish(ctx context.Context, payload []byte) error
}

type publihserService struct {
	pubSub    *gochannel.GoChannel
	topicName string
}

// Publish implements [IPublisherService].
func (p *publihserService) Publish(ctx context.Context, payload []byte) error {
	err := p.pubSub.Publish(
		p.topicName,
		message.NewMessage(watermill.NewUUID(), payload),
	)

	if err != nil {
		return err
	}

	return nil
}

func NewPublisherService(pubSub *gochannel.GoChannel, topicName string) IPublisherService {
	return &publihserService{
		pubSub:    pubSub,
		topicName: topicName,
	}
}
