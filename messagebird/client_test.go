package messagebird

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/epels/birdbroker-go"
)

func TestSendMessage(t *testing.T) {
	t.Run("Bad response", func(t *testing.T) {
		var called bool
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true

			w.WriteHeader(http.StatusForbidden)
		}))

		c := NewClient("")
		c.baseURL = ts.URL

		err := c.SendMessage(context.Background(), &birdbroker.Message{
			Body:       "Hello",
			Originator: "Foo Inc",
			Recipient:  "31612345678",
		})
		if err == nil {
			t.Errorf("Got nil, expected error")
		}
		if !called {
			t.Errorf("Got false, expected true")
		}
	})

	t.Run("OK", func(t *testing.T) {
		var called bool
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true

			if ct := r.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Got %q, expected application/json", ct)
			}
			if auth := r.Header.Get("Authorization"); auth != "AccessKey Secret" {
				t.Errorf("Got %q, expected AccessKey Secret", auth)
			}
			if r.Method != http.MethodPost {
				t.Errorf("Got %q, expected POST", r.Method)
			}
			if r.URL.Path != "/messages" {
				t.Errorf("Got %q, expected /messages", r.URL.Path)
			}

			var data struct {
				Body       string
				Originator string
				Recipients string
			}
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				t.Fatalf("encoding/json: Decoder.Decode: %s", err)
			}
			if data.Body != "Hello" {
				t.Errorf("Got %q, expected Hello", data.Body)
			}
			if data.Originator != "Foo Inc" {
				t.Errorf("Got %q, expected Foo Inc", data.Body)
			}
			if data.Recipients != "31612345678" {
				t.Errorf("Got %q, expected 31612345678", data.Recipients[0])
			}

			w.WriteHeader(http.StatusCreated)
		}))

		c := NewClient("Secret")
		c.baseURL = ts.URL

		err := c.SendMessage(context.Background(), &birdbroker.Message{
			Body:       "Hello",
			Originator: "Foo Inc",
			Recipient:  "31612345678",
		})
		if err != nil {
			t.Errorf("Client: Send: %s", err)
		}
		if !called {
			t.Errorf("Got false, expected true")
		}
	})
}
