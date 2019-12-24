package service

import (
	"context"
	"errors"
	"testing"

	"github.com/epels/birdbroker-go"
	"github.com/epels/birdbroker-go/internal/mock"
)

func TestSendMessage(t *testing.T) {
	t.Run("Validation errors", func(t *testing.T) {
		var s service

		tt := []struct {
			name string
			m    *birdbroker.Message
		}{
			{
				"Body",
				&birdbroker.Message{
					Body:       "",
					Originator: "Foo",
					Recipient:  "31612345678",
				},
			},
			{
				"Originator: empty",
				&birdbroker.Message{
					Body:       "Hello",
					Originator: "",
					Recipient:  "31612345678",
				},
			},
			{
				"Originator: too long",
				&birdbroker.Message{
					Body:       "",
					Originator: "TWELVECHARSS",
					Recipient:  "31612345678",
				},
			},
			{
				"Recipient",
				&birdbroker.Message{
					Body:       "Hello",
					Originator: "Foo",
					Recipient:  "",
				},
			},
		}
		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				var ce birdbroker.ClientError
				if err := s.SendMessage(context.Background(), tc.m); !errors.As(err, &ce) {
					t.Errorf("Got %T, expected ClientError", err)
				}
			})
		}
	})

	t.Run("OK", func(t *testing.T) {
		var called bool
		s := service{
			snd: &mock.MessageQueue{
				PublishFunc: func(m *birdbroker.Message) error {
					called = true

					if m.Body != "Hello" {
						t.Errorf("Got %q, expected Hello", m.Body)
					}
					if m.Originator != "Foo" {
						t.Errorf("Got %q, expected Foo", m.Originator)
					}
					if m.Recipient != "31612345678" {
						t.Errorf("Got %q, expected 31612345678", m.Recipient)
					}

					return nil
				},
			},
		}

		m := birdbroker.Message{
			Body:       "Hello",
			Originator: "Foo",
			Recipient:  "31612345678",
		}
		if err := s.SendMessage(context.Background(), &m); err != nil {
			t.Fatalf("SendMessage: %s", err)
		}
		if !called {
			t.Errorf("Got false, expected true")
		}
	})
}
