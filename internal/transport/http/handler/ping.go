package handler

import (
	"context"
	"net/http"
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
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
