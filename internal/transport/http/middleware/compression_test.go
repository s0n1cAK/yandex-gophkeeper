package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func ungzipBody(t *testing.T, b []byte) []byte {
	t.Helper()
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer zr.Close()
	out, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("read gz: %v", err)
	}
	return out
}

func TestGzipCompression_ResponseCompressed_WhenClientSupportsGzip(t *testing.T) {
	h := GzipCompession()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d want=%d", rr.Code, http.StatusOK)
	}

	if got := rr.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding=%q want gzip", got)
	}
	if vary := rr.Header().Get("Vary"); !strings.Contains(vary, "Accept-Encoding") {
		t.Fatalf("Vary=%q must contain Accept-Encoding", vary)
	}

	plain := ungzipBody(t, rr.Body.Bytes())
	if string(plain) != "hello" {
		t.Fatalf("body=%q want=%q", string(plain), "hello")
	}
}

func TestGzipCompression_ResponseNotCompressed_On401(t *testing.T) {
	h := GzipCompession()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d want=%d", rr.Code, http.StatusUnauthorized)
	}
	if got := rr.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("Content-Encoding=%q want empty", got)
	}
	if strings.Contains(rr.Body.String(), "\x1f\x8b") {
		t.Fatalf("body looks gzipped unexpectedly")
	}
}

func TestGzipCompression_RequestBodyDecompressed_WhenSentGzip(t *testing.T) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte("hello"))
	_ = zw.Close()

	h := GzipCompession()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		_, _ = w.Write(b)
	}))

	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d want=%d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != "hello" {
		t.Fatalf("body=%q want=%q", rr.Body.String(), "hello")
	}
}
