package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func adminHandler(t *testing.T, admins []string, claims *Claims) *httptest.ResponseRecorder {
	t.Helper()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/features", nil)
	if claims != nil {
		req = req.WithContext(context.WithValue(req.Context(), ClaimsKey, claims))
	}
	rec := httptest.NewRecorder()
	AdminRequired(admins)(next).ServeHTTP(rec, req)
	return rec
}

func TestAdminRequiredAllowsListedEmail(t *testing.T) {
	rec := adminHandler(t, []string{"boss@example.com"}, &Claims{Email: "boss@example.com"})
	if rec.Code != http.StatusOK {
		t.Fatalf("listed admin should pass, got %d", rec.Code)
	}
}

func TestAdminRequiredIsCaseAndSpaceInsensitive(t *testing.T) {
	rec := adminHandler(t, []string{"  BOSS@Example.COM "}, &Claims{Email: "boss@example.com"})
	if rec.Code != http.StatusOK {
		t.Fatalf("email match should normalize, got %d", rec.Code)
	}
}

func TestAdminRequiredRejectsUnlistedEmail(t *testing.T) {
	rec := adminHandler(t, []string{"boss@example.com"}, &Claims{Email: "someone@example.com"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("unlisted user should get 403, got %d", rec.Code)
	}
}

// An unconfigured allowlist must fail closed, not open.
func TestAdminRequiredDeniesWhenAllowlistEmpty(t *testing.T) {
	rec := adminHandler(t, nil, &Claims{Email: "boss@example.com"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("empty allowlist must deny everyone, got %d", rec.Code)
	}
}

func TestAdminRequiredRejectsMissingClaims(t *testing.T) {
	rec := adminHandler(t, []string{"boss@example.com"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing claims should get 401, got %d", rec.Code)
	}
}

// Blank entries in ADMIN_EMAILS must not become a wildcard for blank-email tokens.
func TestAdminRequiredIgnoresBlankAllowlistEntries(t *testing.T) {
	rec := adminHandler(t, []string{"", "   "}, &Claims{Email: ""})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("blank entries must not authorize, got %d", rec.Code)
	}
}
