package mock

import (
	"context"

	"github.com/epels/birdbroker-go"
)

type Sender struct {
	SendFunc func(m *birdbroker.Message) error
}

func (s *Sender) Send(ctx context.Context, m *birdbroker.Message) error {
	return s.SendFunc(m)
}
