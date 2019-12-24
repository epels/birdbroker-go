package queue

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/epels/birdbroker-go"
	"github.com/epels/birdbroker-go/internal/mock"
)

func TestPublish(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var called bool
		c := mock.Conn{
			PutFunc: func(body []byte, pri uint32, delay, ttr time.Duration) (uint64, error) {
				called = true

				var m struct {
					Body, Originator, Recipient string
				}
				if err := json.Unmarshal(body, &m); err != nil {
					t.Fatalf("encoding/json: Unmarshal: %s", err)
				}

				if m.Body != "Foo" {
					t.Errorf("Got %q, expected Foo", m.Body)
				}
				if m.Originator != "Bar" {
					t.Errorf("Got %q, expected Bar", m.Originator)
				}
				if m.Recipient != "Baz" {
					t.Errorf("Got %q, expected Baz", m.Recipient)
				}

				return 0, nil
			},
		}
		snd := NewSender(&c)

		err := snd.Send(context.Background(), &birdbroker.Message{
			Body:       "Foo",
			Originator: "Bar",
			Recipient:  "Baz",
		})
		if err != nil {
			t.Errorf("Send: %s", err)
		}
		if !called {
			t.Errorf("Got false, expected true")
		}
	})

	t.Run("Put error", func(t *testing.T) {
		c := mock.Conn{
			PutFunc: func(body []byte, pri uint32, delay, ttr time.Duration) (uint64, error) {
				return 0, errors.New("oops")
			},
		}
		snd := NewSender(&c)

		err := snd.Send(context.Background(), &birdbroker.Message{
			Body:       "Foo",
			Originator: "Bar",
			Recipient:  "Baz",
		})
		if err == nil {
			t.Errorf("Got nil, expected error")
		}
	})
}
