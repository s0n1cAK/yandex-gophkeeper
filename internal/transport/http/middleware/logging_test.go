package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggingMiddleware_LogsMethodURIStatusSize(t *testing.T) {
	core, obs := observer.New(zap.InfoLevel)
	log := zap.New(core)

	h := Logging(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/user/secrets/1/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("code=%d want=%d", rr.Code, http.StatusCreated)
	}

	entries := obs.All()
	if len(entries) != 1 {
		t.Fatalf("log entries=%d want=1", len(entries))
	}

	e := entries[0]
	if e.Message != "http" {
		t.Fatalf("message=%q want=%q", e.Message, "http")
	}

	ctx := e.ContextMap()
	if ctx["method"] != http.MethodPost {
		t.Fatalf("method=%v want=%v", ctx["method"], http.MethodPost)
	}
	if ctx["uri"] != "/api/user/secrets/1/" {
		t.Fatalf("uri=%v", ctx["uri"])
	}
	if ctx["status"] != int64(http.StatusCreated) {
		t.Fatalf("status=%v want=%v", ctx["status"], int64(http.StatusCreated))
	}
	if ctx["size"] != int64(5) {
		t.Fatalf("size=%v want=%v", ctx["size"], int64(5))
	}
	if _, ok := ctx["dur"]; !ok {
		t.Fatalf("expected dur field")
	}
}
