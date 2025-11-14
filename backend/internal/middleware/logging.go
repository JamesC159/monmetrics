package middleware

import (
	"log"
	"net/http"
	"time"
)

// RequestLogger logs HTTP requests
func RequestLogger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap ResponseWriter to capture status code
			wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapper, r)

			log.Printf(
				"%s %s %d %v %s",
				r.Method,
				r.URL.Path,
				wrapper.statusCode,
				time.Since(start),
				getClientIP(r),
			)
		})
	}
}

// responseWrapper wraps http.ResponseWriter to capture status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
