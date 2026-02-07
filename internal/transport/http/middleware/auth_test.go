package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"yandex-gophkeeper/internal/domain"
)

type fakeVerifier struct {
	ok bool
	id int64
}

func (v fakeVerifier) Verify(token string) (int64, bool) {
	if token == "good" && v.ok {
		return v.id, true
	}
	return 0, false
}

func TestAuthMiddleware_SetsUserID(t *testing.T) {
	mw := Auth(fakeVerifier{ok: true, id: 42})

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := UserID(r.Context())
		if !ok || id != domain.UserID(42) {
			t.Fatalf("user id not set")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer good")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
}

func TestAuthMiddleware_MissingToken_401(t *testing.T) {
	mw := Auth(fakeVerifier{ok: true, id: 1})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d want 401", rr.Code)
	}
}
