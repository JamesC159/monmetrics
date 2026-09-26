package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Marketplace providers
const (
	ProviderEbay      = "ebay"
	ProviderTCGPlayer = "tcgplayer"
)

// LinkedAccount is a user's connection to an external marketplace. Tokens are
// stored encrypted and never serialized to clients.
type LinkedAccount struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID           primitive.ObjectID `bson:"user_id" json:"-"`
	Provider         string             `bson:"provider" json:"provider"`
	ExternalUserID   string             `bson:"external_user_id" json:"external_user_id"`
	ExternalUsername string             `bson:"external_username" json:"external_username"`
	AccessTokenEnc   string             `bson:"access_token_enc" json:"-"`
	RefreshTokenEnc  string             `bson:"refresh_token_enc,omitempty" json:"-"`
	Scopes           []string           `bson:"scopes" json:"scopes"`
	ExpiresAt        time.Time          `bson:"expires_at" json:"expires_at"`
	Status           string             `bson:"status" json:"status"` // active, expired, revoked
	LinkedAt         time.Time          `bson:"linked_at" json:"linked_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

// LinkedAccountStatus describes a provider's link state for a user
type LinkedAccountStatus struct {
	Provider    string         `json:"provider"`
	DisplayName string         `json:"display_name"`
	Linked      bool           `json:"linked"`
	Account     *LinkedAccount `json:"account,omitempty"`
	Mock        bool           `json:"mock"`
}

// SoldListing is one completed sale used as a pricing comparable
type SoldListing struct {
	Title     string    `json:"title"`
	Price     float64   `json:"price"`
	Shipping  float64   `json:"shipping"`
	Condition string    `json:"condition"`
	SoldAt    time.Time `json:"sold_at"`
	URL       string    `json:"url"`
	Source    string    `json:"source"`
}

// CompsStats summarizes sold comparables
type CompsStats struct {
	Count          int     `json:"count"`
	Average        float64 `json:"average"`
	Median         float64 `json:"median"`
	Min            float64 `json:"min"`
	Max            float64 `json:"max"`
	SuggestedPrice float64 `json:"suggested_price"`
}

// CompsResponse bundles sold comparables with their stats
type CompsResponse struct {
	Query string        `json:"query"`
	Sold  []SoldListing `json:"sold"`
	Stats CompsStats    `json:"stats"`
}

// ItemSpecific is a provider attribute name/value pair (e.g. "Grade": "10")
type ItemSpecific struct {
	Name  string `bson:"name" json:"name"`
	Value string `bson:"value" json:"value"`
}

// ListingDraft is the editable listing the user reviews before publishing
type ListingDraft struct {
	PortfolioItemID string         `json:"portfolio_item_id"`
	Provider        string         `json:"provider"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	Price           float64        `json:"price"`
	Quantity        int            `json:"quantity"`
	MaxQuantity     int            `json:"max_quantity,omitempty"`
	Condition       string         `json:"condition"`
	CategoryID      string         `json:"category_id"`
	Format          string         `json:"format"` // fixed_price, auction
	Duration        string         `json:"duration"`
	ShippingPrice   float64        `json:"shipping_price"`
	ImageURLs       []string       `json:"image_urls"`
	ItemSpecifics   []ItemSpecific `json:"item_specifics"`
}

// PrefillResponse is returned by the listing prefill endpoint
type PrefillResponse struct {
	Draft ListingDraft  `json:"draft"`
	Comps CompsResponse `json:"comps"`
}

// MarketplaceListing records a listing created on an external marketplace
type MarketplaceListing struct {
	ID                primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID            primitive.ObjectID  `bson:"user_id" json:"-"`
	PortfolioItemID   primitive.ObjectID  `bson:"portfolio_item_id" json:"portfolio_item_id"`
	CardID            *primitive.ObjectID `bson:"card_id,omitempty" json:"card_id,omitempty"`
	Provider          string              `bson:"provider" json:"provider"`
	ExternalListingID string              `bson:"external_listing_id,omitempty" json:"external_listing_id,omitempty"`
	Status            string              `bson:"status" json:"status"` // draft, published, failed, ended
	Title             string              `bson:"title" json:"title"`
	Description       string              `bson:"description" json:"description"`
	Price             float64             `bson:"price" json:"price"`
	Quantity          int                 `bson:"quantity" json:"quantity"`
	Condition         string              `bson:"condition" json:"condition"`
	CategoryID        string              `bson:"category_id,omitempty" json:"category_id,omitempty"`
	Format            string              `bson:"format" json:"format"`
	Duration          string              `bson:"duration,omitempty" json:"duration,omitempty"`
	ShippingPrice     float64             `bson:"shipping_price" json:"shipping_price"`
	ImageURLs         []string            `bson:"image_urls" json:"image_urls"`
	ItemSpecifics     []ItemSpecific      `bson:"item_specifics" json:"item_specifics"`
	ListingURL        string              `bson:"listing_url,omitempty" json:"listing_url,omitempty"`
	ErrorMessage      string              `bson:"error_message,omitempty" json:"error_message,omitempty"`
	CreatedAt         time.Time           `bson:"created_at" json:"created_at"`
	PublishedAt       *time.Time          `bson:"published_at,omitempty" json:"published_at,omitempty"`
	UpdatedAt         time.Time           `bson:"updated_at" json:"updated_at"`
}

// MarketplaceCallbackRequest is sent by the frontend after the OAuth redirect
type MarketplaceCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// PrefillRequest asks the backend to build a listing draft for an owned item
type PrefillRequest struct {
	PortfolioItemID string `json:"portfolio_item_id"`
	Provider        string `json:"provider"`
}

// CreateListingRequest wraps a user-edited draft; publish=false saves a draft only
type CreateListingRequest struct {
	Draft   ListingDraft `json:"draft"`
	Publish bool         `json:"publish"`
}
