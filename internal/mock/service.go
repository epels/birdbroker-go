package mock

import "github.com/epels/birdbroker-go"

type Service struct {
	SendMessageFunc func(*birdbroker.Message) error
}

func (s *Service) SendMessage(m *birdbroker.Message) error {
	return s.SendMessageFunc(m)
}
