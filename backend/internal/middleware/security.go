package middleware

import (
	"net/http"
)

// SecurityHeaders implements OWASP security headers
func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent XSS attacks
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// HSTS for HTTPS
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			// Content Security Policy
			csp := "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data: https:; " +
				"font-src 'self'; " +
				"connect-src 'self'; " +
				"media-src 'self'; " +
				"object-src 'none'; " +
				"child-src 'self'; " +
				"frame-ancestors 'none'; " +
				"form-action 'self'; " +
				"base-uri 'self';"
			w.Header().Set("Content-Security-Policy", csp)

			// Referrer Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Feature Policy
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			next.ServeHTTP(w, r)
		})
	}
}
