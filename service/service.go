package service

import (
	"fmt"

	"github.com/epels/birdbroker-go"
)

type service struct {
	mq messageQueue
}

type messageQueue interface {
	Publish(m *birdbroker.Message) error
}

func New(mq messageQueue) *service {
	return &service{mq: mq}
}

// SendMessage publishes the message to the message queue for sending. Delivery
// is not guaranteed on err==nil: this only means the message is accepted and
// queued for delivery.
func (s *service) SendMessage(m *birdbroker.Message) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("message: Validate: %w", err)
	}
	if err := s.mq.Publish(m); err != nil {
		return fmt.Errorf("%T: Publish: %s", s.mq, err)
	}
	return nil
}
