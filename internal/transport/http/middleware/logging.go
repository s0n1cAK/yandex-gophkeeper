package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type rw struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *rw) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *rw) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.size += n
	return n, err
}

func Logging(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &rw{ResponseWriter: w}
			next.ServeHTTP(ww, r)

			log.Info("http",
				zap.String("method", r.Method),
				zap.String("uri", r.URL.Path),
				zap.Int("status", ww.status),
				zap.Int("size", ww.size),
				zap.Duration("dur", time.Since(start)),
			)
		})
	}
}
