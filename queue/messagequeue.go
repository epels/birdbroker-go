package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/epels/birdbroker-go"
)

const defaultPriority = 0

type messageQueue struct {
	conn conn
}

type conn interface {
	Put(body []byte, pri uint32, delay, ttr time.Duration) (uint64, error)
}

func NewMessageQueue(c conn) *messageQueue {
	return &messageQueue{conn: c}
}

func (mq *messageQueue) Publish(m *birdbroker.Message) error {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("encoding/json: Marshal: %s", err)
	}

	_, err = mq.conn.Put(b, defaultPriority, 0*time.Second, 1*time.Minute)
	if err != nil {
		return fmt.Errorf("%T: Put: %s", mq.conn, err)
	}
	return nil
}
