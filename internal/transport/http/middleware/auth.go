package middleware

import (
	"context"
	"net/http"
	"strings"
	"yandex-gophkeeper/internal/domain"
	"yandex-gophkeeper/internal/transport/http/respond"
)

type ctxKey string

const ctxUserID ctxKey = "userID"

type TokenVerifier interface {
	Verify(token string) (userID int64, ok bool)
}

func Auth(v TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				respond.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			uid, ok := v.Verify(token)
			if !ok {
				respond.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := WithUserID(r.Context(), domain.UserID(uid))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) string {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if h == "" {
		return ""
	}
	const pref = "Bearer "
	if !strings.HasPrefix(h, pref) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, pref))
}

func WithUserID(ctx context.Context, id domain.UserID) context.Context {
	return context.WithValue(ctx, ctxUserID, id)
}

func UserID(ctx context.Context) (domain.UserID, bool) {
	v := ctx.Value(ctxUserID)
	id, ok := v.(domain.UserID)
	return id, ok
}
