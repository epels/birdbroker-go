package service

import (
	"fmt"

	"github.com/epels/birdbroker-go"
)

type service struct {
	snd sender
}

type sender interface {
	Send(m *birdbroker.Message) error
}

func New(snd sender) *service {
	return &service{snd: snd}
}

func (s *service) SendMessage(m *birdbroker.Message) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("message: Validate: %w", err)
	}
	if err := s.snd.Send(m); err != nil {
		return fmt.Errorf("%T: Send: %s", s.snd, err)
	}
	return nil
}
