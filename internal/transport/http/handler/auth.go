package handler

import (
	"context"
	"net/http"
	"strings"
	"yandex-gophkeeper/internal/domain"
	"yandex-gophkeeper/internal/transport/http/middleware"

	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (token string, err error)
	Login(ctx context.Context, username, password string) (token string, err error)
}

type AuthHandler struct {
	log *zap.Logger
	svc AuthService
}

func NewAuth(log *zap.Logger, svc AuthService) *AuthHandler {
	return &AuthHandler{log: log, svc: svc}
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func userIDFromContext(ctx context.Context) (domain.UserID, bool) {
	return middleware.UserID(ctx)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password required")
		return
	}

	token, err := h.svc.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		h.log.Error("register failed", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password required")
		return
	}

	token, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		h.log.Warn("login failed", zap.Error(err))
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
