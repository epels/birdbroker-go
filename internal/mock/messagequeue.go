package mock

import (
	"context"

	"github.com/epels/birdbroker-go"
)

type MessageQueue struct {
	PublishFunc func(m *birdbroker.Message) error
}

func (mq *MessageQueue) Send(ctx context.Context, m *birdbroker.Message) error {
	return mq.PublishFunc(m)
}
