package server

import (
	"bytes"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           new(bytes.Buffer),
	}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	lrw.body.Write(b)
	return lrw.ResponseWriter.Write(b)
}

func LoggerMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var requestBody []byte
			if r.Body != nil {
				requestBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			logger.Info("httpserver: incoming request",
				zap.String("url", r.Method+": "+r.URL.String()),
				zap.String("requestBody", string(requestBody)),
			)

			lrw := newLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r)

			duration := time.Since(start)
			logger.Info("httpserver: request processed",
				zap.String("url", r.Method+": "+r.URL.String()),
				zap.String("requestBody", string(requestBody)),
				zap.Int("StatusResponse", lrw.statusCode),
				zap.Duration("Duration", duration),
			)
		})
	}
}
