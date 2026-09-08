package handler

import (
	"context"
	"net/http"
	"yandex-gophkeeper/internal/transport/http/respond"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type PingHandler struct {
	pinger Pinger
}

func NewPing(p Pinger) *PingHandler {
	return &PingHandler{pinger: p}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.pinger != nil {
		if err := h.pinger.Ping(r.Context()); err != nil {
			respond.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	respond.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
