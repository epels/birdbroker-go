package mock

import "github.com/epels/birdbroker-go"

type MessageQueue struct {
	PublishFunc func(m *birdbroker.Message) error
}

func (mq *MessageQueue) Publish(m *birdbroker.Message) error {
	return mq.PublishFunc(m)
}
