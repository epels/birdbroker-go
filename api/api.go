package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"

	"github.com/epels/birdbroker-go"
)

type handler struct {
	http.Handler
	handlerOnce sync.Once // Guards initialization of Handler.

	svc service
}

type service interface {
	SendMessage(m *birdbroker.Message) error
}

func NewHandler(s service) *handler {
	h := &handler{svc: s}
	return h.withRoutes()
}

func (h *handler) withRoutes() *handler {
	h.handlerOnce.Do(func() {
		r := mux.NewRouter()
		r.Use(h.logMiddleware)
		r.HandleFunc("/messages", h.sendMessage).Methods(http.MethodPost)
		h.Handler = r
	})
	return h
}

func (h *handler) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

func (h *handler) error(w http.ResponseWriter, err error) {
	type res struct {
		Error string `json:"error"`
	}

	var ce birdbroker.ClientError
	if !errors.As(err, &ce) {
		h.response(w, http.StatusInternalServerError, res{
			Error: http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	h.response(w, http.StatusBadRequest, res{
		Error: ce.Error(),
	})
}

func (h *handler) response(w http.ResponseWriter, code int, v interface{}) {
	w.WriteHeader(code)

	if v != nil {
		b, err := json.Marshal(v)
		if err != nil {
			log.Printf("encoding/json: Marshal: %s", err)
		}
		if _, err = w.Write(b); err != nil {
			log.Printf("%T: Write: %s", w, err)
		}
	}
}

func (h *handler) sendMessage(w http.ResponseWriter, r *http.Request) {
	var m birdbroker.Message
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.error(w, birdbroker.ClientError{
			Reason: "Cannot decode request body",
		})
		return
	}

	if err := h.svc.SendMessage(&m); err != nil {
		log.Printf("%T: SendMessage: %s", h.svc, err)
		h.error(w, err)
		return
	}

	h.response(w, http.StatusCreated, nil)
}
