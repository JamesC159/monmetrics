package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jamesc159/monmetrics/internal/features"
	"github.com/jamesc159/monmetrics/internal/middleware"
)

// GetFeatures returns the effective flag map so the frontend can hide gated UI.
// Public: knowing which surfaces are disabled is not sensitive, and the client needs it
// before authenticating.
func (h *Handlers) GetFeatures(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, h.features.Public())
}

// GetAdminFeatures returns every flag with its definition and provenance.
func (h *Handlers) GetAdminFeatures(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, h.features.States())
}

type updateFeatureRequest struct {
	Enabled bool `json:"enabled"`
}

// UpdateAdminFeature toggles one flag. Takes effect immediately - no redeploy.
func (h *Handlers) UpdateAdminFeature(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if !features.IsKnown(key) {
		h.sendError(w, "Unknown feature flag", http.StatusNotFound, nil)
		return
	}

	var req updateFeatureRequest
	if err := decodeJSON(w, r, &req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest, nil)
		return
	}

	actor := ""
	if claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims); ok && claims != nil {
		actor = claims.Email
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if err := h.features.Set(ctx, key, req.Enabled, actor); err != nil {
		h.sendError(w, "Could not update feature flag", http.StatusInternalServerError, nil)
		return
	}

	sendJSON(w, http.StatusOK, h.features.States())
}
