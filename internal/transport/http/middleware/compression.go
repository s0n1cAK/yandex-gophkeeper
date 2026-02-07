package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"yandex-gophkeeper/internal/transport/http/respond"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
		wroteHeader  bool
	}

	compressWriter struct {
		w           http.ResponseWriter
		zw          *gzip.Writer
		enabled     bool
		wroteHeader bool
	}

	compressReader struct {
		r  io.ReadCloser
		zr *gzip.Reader
	}
)

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{w: w}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}
	if c.enabled {
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true

	if statusCode < 300 {
		c.enabled = true
		c.w.Header().Set("Content-Encoding", "gzip")
		c.zw = gzip.NewWriter(c.w)
	}

	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	if c.enabled && c.zw != nil {
		return c.zw.Close()
	}
	return nil
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipCompession() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		compressFn := func(w http.ResponseWriter, r *http.Request) {

			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				w.Header().Add("Vary", "Accept-Encoding")

				cw := newCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					respond.WriteError(w, http.StatusBadRequest, "bad gzip body")
					return
				}
				r.Body = cr
				defer cr.Close()
			}
			h.ServeHTTP(ow, r)

		}
		return http.HandlerFunc(compressFn)
	}
}
