// Package marketplace abstracts external marketplaces (eBay, TCGPlayer) behind
// provider interfaces so mock and real implementations are interchangeable.
package marketplace

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jamesc159/monmetrics/internal/models"
)

// Tokens are OAuth credentials returned by a provider.
type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Scopes       []string
}

// Account identifies the external marketplace account.
type Account struct {
	ExternalUserID string
	Username       string
}

// ListingResult is returned after a listing is published.
type ListingResult struct {
	ExternalID string
	URL        string
}

// Provider handles account linking and listing creation for one marketplace.
type Provider interface {
	Name() string
	DisplayName() string
	AuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (Tokens, Account, error)
	RefreshTokens(ctx context.Context, refreshToken string) (Tokens, error)
	CreateListing(ctx context.Context, accessToken string, draft models.ListingDraft) (ListingResult, error)
}

// CompsQuery describes the item to find sold comparables for.
type CompsQuery struct {
	CardID    *primitive.ObjectID
	Keywords  string
	ItemType  string
	Condition string
	Grading   *models.Grading
	Source    string
	Days      int
}

// CompsProvider searches completed/sold listings.
type CompsProvider interface {
	SearchSold(ctx context.Context, q CompsQuery) ([]models.SoldListing, error)
}

// Registry resolves providers by name.
type Registry struct {
	Mode      string
	providers map[string]Provider
	comps     CompsProvider
}

// NewRegistry builds the provider set for the configured mode.
func NewRegistry(mode string, db *mongo.Database, apiBaseURL string) (*Registry, error) {
	switch mode {
	case "mock":
		return &Registry{
			Mode: mode,
			providers: map[string]Provider{
				models.ProviderEbay:      newMockProvider(models.ProviderEbay, "eBay", apiBaseURL),
				models.ProviderTCGPlayer: newMockProvider(models.ProviderTCGPlayer, "TCGPlayer", apiBaseURL),
			},
			comps: &mockComps{db: db},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported MARKETPLACE_MODE %q (only \"mock\" is implemented)", mode)
	}
}

// Provider returns the named provider.
func (r *Registry) Provider(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// Providers returns providers in stable display order.
func (r *Registry) Providers() []Provider {
	out := make([]Provider, 0, len(r.providers))
	for _, name := range []string{models.ProviderEbay, models.ProviderTCGPlayer} {
		if p, ok := r.providers[name]; ok {
			out = append(out, p)
		}
	}
	return out
}

// Comps returns the sold-listings provider.
func (r *Registry) Comps() CompsProvider {
	return r.comps
}

// IsMock reports whether mock providers are active.
func (r *Registry) IsMock() bool {
	return r.Mode == "mock"
}
