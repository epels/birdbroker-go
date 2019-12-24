package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/epels/birdbroker-go"
)

const defaultPriority = 0

type sender struct {
	conn conn
}

type conn interface {
	Put(body []byte, pri uint32, delay, ttr time.Duration) (uint64, error)
}

func NewSender(c conn) *sender {
	return &sender{conn: c}
}

func (snd *sender) Send(ctx context.Context, m *birdbroker.Message) error {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("encoding/json: Marshal: %s", err)
	}

	_, err = snd.conn.Put(b, defaultPriority, 0*time.Second, 1*time.Minute)
	if err != nil {
		return fmt.Errorf("%T: Put: %s", snd.conn, err)
	}
	return nil
}
