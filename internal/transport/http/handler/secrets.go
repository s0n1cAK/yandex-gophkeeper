package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"yandex-gophkeeper/internal/domain"
	secretsvc "yandex-gophkeeper/internal/service/secrets"
	"yandex-gophkeeper/internal/transport/http/respond"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type SecretsService interface {
	Create(ctx context.Context, ownerID domain.UserID, in secretsvc.SecretUpsert) (domain.SecretID, error)
	List(ctx context.Context, ownerID domain.UserID) ([]domain.SecretMeta, error)
	Get(ctx context.Context, ownerID domain.UserID, id domain.SecretID) (domain.Secret, []byte, error)
	Update(ctx context.Context, ownerID domain.UserID, id domain.SecretID, in secretsvc.SecretUpsert) error
	Delete(ctx context.Context, ownerID domain.UserID, id domain.SecretID) error
}

type secretUpsertRequest struct {
	Type    domain.SecretType `json:"type"`
	Comment string            `json:"comment,omitempty"`
	Data    string            `json:"data"`
}

type SecretsHandler struct {
	log *zap.Logger
	svc SecretsService
}

func NewSecrets(log *zap.Logger, svc SecretsService) *SecretsHandler {
	return &SecretsHandler{log: log, svc: svc}
}

func (h *SecretsHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFromContext(r.Context())
	if !ok {
		respond.WriteError(w, http.StatusUnauthorized, "missing auth")
		return
	}

	items, err := h.svc.List(r.Context(), uid)
	if err != nil {
		h.log.Error("list secrets failed", zap.Error(err))
		respond.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respond.WriteJSON(w, http.StatusOK, map[string]any{"secrets": items})
}

func (h *SecretsHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFromContext(r.Context())
	if !ok {
		respond.WriteError(w, http.StatusUnauthorized, "missing auth")
		return
	}

	var req secretUpsertRequest
	if err := respond.DecodeJSON(r, &req); err != nil {
		respond.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	upsert, err := toUpsert(req)
	if err != nil {
		respond.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.svc.Create(r.Context(), uid, upsert)
	if err != nil {
		h.log.Error("create secret failed", zap.Error(err))
		respond.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	respond.WriteJSON(w, http.StatusOK, map[string]any{"id": id})
}

func (h *SecretsHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFromContext(r.Context())
	if !ok {
		respond.WriteError(w, http.StatusUnauthorized, "missing auth")
		return
	}

	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		respond.WriteError(w, http.StatusBadRequest, "bad id")
		return
	}

	sec, plain, err := h.svc.Get(r.Context(), uid, domain.SecretID(id))
	if err != nil {
		h.log.Warn("get secret failed", zap.Error(err))
		respond.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	data := encodePlainForResponse(sec.Type, plain)

	respond.WriteJSON(w, http.StatusOK, map[string]any{
		"secret": sec,
		"data":   data,
	})
}

func (h *SecretsHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFromContext(r.Context())
	if !ok {
		respond.WriteError(w, http.StatusUnauthorized, "missing auth")
		return
	}

	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		respond.WriteError(w, http.StatusBadRequest, "bad id")
		return
	}

	var req secretUpsertRequest
	if err := respond.DecodeJSON(r, &req); err != nil {
		respond.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	upsert, err := toUpsert(req)
	if err != nil {
		respond.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.Update(r.Context(), uid, domain.SecretID(id), upsert); err != nil {
		h.log.Error("update secret failed", zap.Error(err))
		respond.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	respond.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *SecretsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := userIDFromContext(r.Context())
	if !ok {
		respond.WriteError(w, http.StatusUnauthorized, "missing auth")
		return
	}

	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		respond.WriteError(w, http.StatusBadRequest, "bad id")
		return
	}

	if err := h.svc.Delete(r.Context(), uid, domain.SecretID(id)); err != nil {
		h.log.Warn("delete secret failed", zap.Error(err))
		respond.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	respond.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func toUpsert(req secretUpsertRequest) (secretsvc.SecretUpsert, error) {
	req.Comment = strings.TrimSpace(req.Comment)

	if err := domain.ValidateSecretType(req.Type); err != nil {
		return secretsvc.SecretUpsert{}, err
	}

	var data []byte

	switch req.Type {
	case domain.SecretBinary:
		b, err := base64.StdEncoding.DecodeString(req.Data)
		if err != nil {
			return secretsvc.SecretUpsert{}, domain.ErrInvalidPayload
		}
		data = b
	case domain.SecretBankCard:
		var p domain.BankCardPayload
		if err := json.Unmarshal([]byte(req.Data), &p); err != nil {
			return secretsvc.SecretUpsert{}, domain.ErrInvalidPayload
		}
		if err := p.Validate(); err != nil {
			return secretsvc.SecretUpsert{}, err
		}
		p.Number = p.NormalizeNumber()
		b, err := json.Marshal(p)
		if err != nil {
			return secretsvc.SecretUpsert{}, domain.ErrInternal
		}
		data = b

	default:
		data = []byte(req.Data)
	}

	return secretsvc.SecretUpsert{
		Type:    req.Type,
		Comment: req.Comment,
		Data:    data,
	}, nil
}

func encodePlainForResponse(t domain.SecretType, plain []byte) string {
	if t == domain.SecretBinary {
		return base64.StdEncoding.EncodeToString(plain)
	}
	return string(plain)
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}
