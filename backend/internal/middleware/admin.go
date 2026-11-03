package middleware

import (
	"net/http"
	"strings"
)

// AdminRequired gates admin-only routes. It must be chained after AuthRequired, which
// puts the verified JWT claims on the context.
//
// Membership is checked against a server-side allowlist rather than a claim or a DB
// field, so a token cannot assert its own privilege and revocation does not wait for
// the 24h token expiry.
func AdminRequired(adminEmails []string) Middleware {
	allowed := make(map[string]struct{}, len(adminEmails))
	for _, email := range adminEmails {
		if normalized := strings.ToLower(strings.TrimSpace(email)); normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// An empty allowlist denies everyone; a misconfigured deployment must not open up.
			if len(allowed) == 0 {
				http.Error(w, "Admin access is not configured", http.StatusForbidden)
				return
			}

			claims, ok := r.Context().Value(ClaimsKey).(*Claims)
			if !ok || claims == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if _, isAdmin := allowed[strings.ToLower(strings.TrimSpace(claims.Email))]; !isAdmin {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
