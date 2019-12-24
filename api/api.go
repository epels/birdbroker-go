package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/epels/birdbroker-go"
)

type handler struct {
	http.Handler

	svc service
}

type service interface {
	SendMessage(m *birdbroker.Message) error
}

func NewHandler(s service) (*handler, error) {
	return &handler{
		svc: s,
	}, nil
}

func (h *handler) withRoutes() *handler {
	r := mux.NewRouter()
	r.HandleFunc("/messages", h.sendMessage).Methods(http.MethodPost)
	h.Handler = r
	return h
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
		h.error(w, err)
		return
	}

	h.response(w, http.StatusCreated, nil)
}
