// Package features provides runtime-toggleable feature flags.
//
// Flags gate product surfaces whose backing data source does not exist yet. They live in
// Mongo rather than env vars so an operator can enable a surface the moment its ingest
// lands, without a redeploy. Reads are served from an in-memory cache.
package features

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const Collection = "feature_flags"

// Keys are namespaced by the data source they depend on.
const (
	KeyEbaySoldListings = "ebay.sold_listings"
	KeyEbayGrades       = "ebay.grades"
	KeyPriceSummary     = "price.summary_windows"
	KeyMarketplaceComps = "marketplace.comps"
	KeyAllGames         = "catalog.all_games"
)

// Definition describes a flag. Blocker records what has to exist before it can be turned
// on, so the admin UI can explain why a toggle is off rather than just showing a switch.
type Definition struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Default     bool   `json:"default"`
	Blocker     string `json:"blocker,omitempty"`
}

// Definitions is the authoritative list. A flag absent from Mongo falls back to Default.
var Definitions = []Definition{
	{
		Key:         KeyEbaySoldListings,
		Label:       "Recent eBay sales",
		Description: "Per-sale listing table on the card page, backed by the sold_listings collection.",
		Default:     false,
		Blocker:     "Needs an eBay sold-listing source. The HTML scraper is blocked (403) and TCGCSV cannot supply individual sales.",
	},
	{
		Key:         KeyEbayGrades,
		Label:       "Graded price series",
		Description: "Grade selector (PSA/BGS/CGC) on the price chart.",
		Default:     false,
		Blocker:     "Needs a graded-sales source. TCGCSV exposes printings (Normal/Holofoil), not grades.",
	},
	{
		Key:         KeyPriceSummary,
		Label:       "Rolling window stats",
		Description: "1d/7d/30d/90d/1y/5y summary strip on the card page.",
		Default:     false,
		Blocker:     "Needs price_summaries. Rebuildable from TCGCSV daily snapshots once ingest accrues history.",
	},
	{
		Key:         KeyMarketplaceComps,
		Label:       "Marketplace comps",
		Description: "Comparable-sales pricing in the listing composer.",
		Default:     false,
		Blocker:     "Currently synthesises fake comps from the prices collection. Needs a real sold-comp source.",
	},
	{
		Key:         KeyAllGames,
		Label:       "All trading card games",
		Description: "Expose every game in browse and search, not just the ENABLED_GAMES allowlist.",
		Default:     false,
		Blocker:     "Non-Magic games have no price history. Enable once TCGCSV ingest covers their categories.",
	},
}

// Flag is one stored override.
type Flag struct {
	Key       string    `bson:"key" json:"key"`
	Enabled   bool      `bson:"enabled" json:"enabled"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	UpdatedBy string    `bson:"updated_by,omitempty" json:"updated_by,omitempty"`
}

// State is a definition joined with its effective value, for the admin UI.
type State struct {
	Definition
	Enabled   bool       `json:"enabled"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	UpdatedBy string     `json:"updated_by,omitempty"`
}

// Registry caches effective flag values so request paths never hit Mongo.
type Registry struct {
	db *mongo.Database

	mu       sync.RWMutex
	enabled  map[string]bool
	stored   map[string]Flag
	loadedAt time.Time
}

func NewRegistry(db *mongo.Database) *Registry {
	r := &Registry{db: db, enabled: map[string]bool{}, stored: map[string]Flag{}}
	for _, d := range Definitions {
		r.enabled[d.Key] = d.Default
	}
	return r
}

// IsKnown reports whether key is a declared flag. Guards the admin write path against
// writing arbitrary keys into the collection.
func IsKnown(key string) bool {
	for _, d := range Definitions {
		if d.Key == key {
			return true
		}
	}
	return false
}

// Enabled returns the effective value, falling back to the declared default.
func (r *Registry) Enabled(key string) bool {
	if r == nil {
		return defaultFor(key)
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if v, ok := r.enabled[key]; ok {
		return v
	}
	return defaultFor(key)
}

func defaultFor(key string) bool {
	for _, d := range Definitions {
		if d.Key == key {
			return d.Default
		}
	}
	return false
}

// Load refreshes the cache from Mongo. Safe to call on a cold collection.
func (r *Registry) Load(ctx context.Context) error {
	cursor, err := r.db.Collection(Collection).Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var flags []Flag
	if err := cursor.All(ctx, &flags); err != nil {
		return err
	}

	next := make(map[string]bool, len(Definitions))
	for _, d := range Definitions {
		next[d.Key] = d.Default
	}
	stored := make(map[string]Flag, len(flags))
	for _, f := range flags {
		if !IsKnown(f.Key) {
			continue // stale key from a removed feature
		}
		next[f.Key] = f.Enabled
		stored[f.Key] = f
	}

	r.mu.Lock()
	r.enabled, r.stored, r.loadedAt = next, stored, time.Now()
	r.mu.Unlock()
	return nil
}

// Set persists an override and refreshes the cache so the change takes effect immediately.
func (r *Registry) Set(ctx context.Context, key string, enabled bool, actor string) error {
	flag := Flag{Key: key, Enabled: enabled, UpdatedAt: time.Now(), UpdatedBy: actor}
	_, err := r.db.Collection(Collection).UpdateOne(ctx,
		bson.M{"key": key},
		bson.M{"$set": flag},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}
	return r.Load(ctx)
}

// States returns every definition with its effective value, in declaration order.
func (r *Registry) States() []State {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]State, 0, len(Definitions))
	for _, d := range Definitions {
		s := State{Definition: d, Enabled: d.Default}
		if v, ok := r.enabled[d.Key]; ok {
			s.Enabled = v
		}
		if f, ok := r.stored[d.Key]; ok {
			updated := f.UpdatedAt
			s.UpdatedAt = &updated
			s.UpdatedBy = f.UpdatedBy
		}
		out = append(out, s)
	}
	return out
}

// Public is the flag map the frontend uses to hide gated UI.
func (r *Registry) Public() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]bool, len(r.enabled))
	for k, v := range r.enabled {
		out[k] = v
	}
	return out
}
