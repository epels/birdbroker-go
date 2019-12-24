package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/epels/birdbroker-go"
	"github.com/epels/birdbroker-go/internal/mock"
)

func TestError(t *testing.T) {
	t.Run("Client error", func(t *testing.T) {
		rec := httptest.NewRecorder()

		var h handler
		h.error(rec, birdbroker.ClientError{
			Reason: "uh oh",
		})

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Got %d, expected 400", rec.Code)
		}
		if b := rec.Body.String(); b != `{"error":"uh oh"}` {
			t.Errorf(`Got %q, expected {"error":"uh oh"}`, b)
		}
	})

	t.Run("Internal error", func(t *testing.T) {
		rec := httptest.NewRecorder()

		var h handler
		h.error(rec, errors.New("uh oh"))

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Got %d, expected 500", rec.Code)
		}
		if b := rec.Body.String(); b != `{"error":"Internal Server Error"}` {
			t.Errorf(`Got %q, expected {"error":"Internal Server Error"}`, b)
		}
	})
}

func TestResponse(t *testing.T) {
	t.Run("With body", func(t *testing.T) {
		rec := httptest.NewRecorder()

		var h handler
		h.response(rec, http.StatusCreated, struct {
			Message string `json:"message"`
		}{
			Message: "hello",
		})

		if rec.Code != http.StatusCreated {
			t.Errorf("Got %d, expected 201", rec.Code)
		}
		if b := rec.Body.String(); b != `{"message":"hello"}` {
			t.Errorf(`Got %q, expected {"message":"hello"}`, b)
		}
	})

	t.Run("Without body", func(t *testing.T) {
		rec := httptest.NewRecorder()

		var h handler
		h.response(rec, http.StatusCreated, nil)

		if rec.Code != http.StatusCreated {
			t.Errorf("Got %d, expected 201", rec.Code)
		}
		if b := rec.Body.String(); b != "" {
			t.Errorf("Got %q, expected empty string", b)
		}
	})
}

func TestSendMessage(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var called bool
		h := &handler{
			svc: &mock.Service{
				SendMessageFunc: func(m *birdbroker.Message) error {
					called = true

					if m.Body != "Hello!" {
						t.Errorf("Got %q, expected Hello!", m.Body)
					}
					if m.Originator != "Hello" {
						t.Errorf("Got %q, expected Hello", m.Originator)
					}
					if m.Recipient != "31612345678" {
						t.Errorf("Got %q, expected 31612345678", m.Recipient)
					}

					return nil
				},
			},
		}
		h.withRoutes()

		rec := httptest.NewRecorder()
		rr := strings.NewReader(`{
	"body": "Hello!",
	"originator": "Hello",
	"recipient": "31612345678"
}`)
		req := httptest.NewRequest(http.MethodPost, "/messages", rr)

		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Errorf("Got %d, expected 201", rec.Code)
		}
		if !called {
			t.Errorf("Got false, expected true")
		}
	})

	t.Run("Bad request", func(t *testing.T) {
		var called bool
		h := &handler{
			svc: &mock.Service{
				SendMessageFunc: func(m *birdbroker.Message) error {
					called = true
					return birdbroker.ClientError{Reason: "oops"}
				},
			},
		}
		h.withRoutes()

		rec := httptest.NewRecorder()
		rr := strings.NewReader("{}")
		req := httptest.NewRequest(http.MethodPost, "/messages", rr)

		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("Got %d, expected 400", rec.Code)
		}
		if s := rec.Body.String(); s != `{"error":"oops"}` {
			t.Errorf(`Got %q, expected {"error":"oops"}`, s)
		}
		if !called {
			t.Errorf("Got false, expected true")
		}
	})

	t.Run("Internal error", func(t *testing.T) {
		var called bool
		h := &handler{
			svc: &mock.Service{
				SendMessageFunc: func(m *birdbroker.Message) error {
					called = true
					return errors.New("oops")
				},
			},
		}
		h.withRoutes()

		rec := httptest.NewRecorder()
		rr := strings.NewReader("{}")
		req := httptest.NewRequest(http.MethodPost, "/messages", rr)

		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Got %d, expected 500", rec.Code)
		}
		if s := rec.Body.String(); s != `{"error":"Internal Server Error"}` {
			t.Errorf(`Got %q, expected {"error":"Internal Server Error"}`, s)
		}
		if !called {
			t.Errorf("Got false, expected true")
		}
	})
}
