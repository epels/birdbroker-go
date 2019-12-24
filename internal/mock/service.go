package mock

import (
	"context"

	"github.com/epels/birdbroker-go"
)

type Service struct {
	SendMessageFunc func(*birdbroker.Message) error
}

func (s *Service) SendMessage(ctx context.Context, m *birdbroker.Message) error {
	return s.SendMessageFunc(m)
}
