package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		rw := &responseWriter{
			ResponseWriter: w,
			status:         200,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		log.Printf("REQUEST [%s] %s %s %d %d %v",
			requestID,
			r.Method,
			r.URL.Path,
			rw.status,
			rw.size,
			duration,
		)
	})
}
