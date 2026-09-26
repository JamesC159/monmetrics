package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/marketplace"
	"github.com/jamesc159/monmetrics/internal/models"
)

var (
	validListingFormats = map[string]bool{"fixed_price": true, "auction": true}
	validEbayDurations  = map[string]bool{"GTC": true, "DAYS_3": true, "DAYS_5": true, "DAYS_7": true, "DAYS_10": true}
)

func (h *Handlers) providerFromPath(w http.ResponseWriter, r *http.Request) (marketplace.Provider, bool) {
	p, ok := h.market.Provider(r.PathValue("provider"))
	if !ok {
		h.sendError(w, "Unknown marketplace provider", http.StatusNotFound, nil)
		return nil, false
	}
	return p, true
}

// GetLinkedAccounts lists link status for every supported provider
func (h *Handlers) GetLinkedAccounts(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := h.db.Collection("linked_accounts").Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		h.sendError(w, "Error retrieving linked accounts", http.StatusInternalServerError, nil)
		return
	}
	defer cur.Close(ctx)
	var accounts []models.LinkedAccount
	if err := cur.All(ctx, &accounts); err != nil {
		h.sendError(w, "Error decoding linked accounts", http.StatusInternalServerError, nil)
		return
	}
	byProvider := map[string]*models.LinkedAccount{}
	for i := range accounts {
		byProvider[accounts[i].Provider] = &accounts[i]
	}

	out := make([]models.LinkedAccountStatus, 0)
	for _, p := range h.market.Providers() {
		acct := byProvider[p.Name()]
		out = append(out, models.LinkedAccountStatus{
			Provider:    p.Name(),
			DisplayName: p.DisplayName(),
			Linked:      acct != nil && acct.Status != "revoked",
			Account:     acct,
			Mock:        h.market.IsMock(),
		})
	}
	sendJSON(w, http.StatusOK, out)
}

// ConnectMarketplace starts the OAuth flow and returns the provider consent URL
func (h *Handlers) ConnectMarketplace(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	p, ok := h.providerFromPath(w, r)
	if !ok {
		return
	}
	state, err := signOAuthState(h.config.JWTSecret, userID, p.Name(), time.Now())
	if err != nil {
		h.sendError(w, "Could not start account linking", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"auth_url": p.AuthURL(state)})
}

// MockAuthorize simulates a provider consent screen (mock mode only). It is public
// because the browser reaches it via redirect; the signed state authorizes it.
func (h *Handlers) MockAuthorize(w http.ResponseWriter, r *http.Request) {
	if !h.market.IsMock() {
		http.NotFound(w, r)
		return
	}
	p, ok := h.providerFromPath(w, r)
	if !ok {
		return
	}
	state := r.URL.Query().Get("state")
	if _, err := verifyOAuthState(h.config.JWTSecret, state, p.Name(), time.Now()); err != nil {
		h.sendError(w, "Invalid or expired authorization request", http.StatusBadRequest, nil)
		return
	}
	q := url.Values{}
	q.Set("provider", p.Name())
	q.Set("code", marketplace.NewMockCode())
	q.Set("state", state)
	http.Redirect(w, r, h.config.FrontendURL+"/marketplace/callback?"+q.Encode(), http.StatusFound)
}

// MarketplaceCallback completes OAuth: verifies state, exchanges the code, stores encrypted tokens
func (h *Handlers) MarketplaceCallback(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	p, ok := h.providerFromPath(w, r)
	if !ok {
		return
	}
	var req models.MarketplaceCallbackRequest
	if err := decodeJSON(w, r, &req); err != nil || req.Code == "" || len(req.Code) > 512 {
		h.sendError(w, "Invalid request body", http.StatusBadRequest, nil)
		return
	}
	stateUser, err := verifyOAuthState(h.config.JWTSecret, req.State, p.Name(), time.Now())
	if err != nil || stateUser != userID {
		h.sendError(w, "Invalid or expired authorization state", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tokens, account, err := p.ExchangeCode(ctx, req.Code)
	if err != nil {
		h.sendError(w, "Authorization failed", http.StatusBadGateway, nil)
		return
	}
	if err := h.storeLinkedAccount(ctx, userID, p.Name(), tokens, account); err != nil {
		h.sendError(w, "Error saving linked account", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"provider": p.Name(), "username": account.Username})
}

func (h *Handlers) storeLinkedAccount(ctx context.Context, userID primitive.ObjectID, provider string, tokens marketplace.Tokens, account marketplace.Account) error {
	accessEnc, err := h.cipher.Encrypt(tokens.AccessToken)
	if err != nil {
		return err
	}
	refreshEnc := ""
	if tokens.RefreshToken != "" {
		if refreshEnc, err = h.cipher.Encrypt(tokens.RefreshToken); err != nil {
			return err
		}
	}
	now := time.Now().UTC()
	_, err = h.db.Collection("linked_accounts").UpdateOne(ctx,
		bson.M{"user_id": userID, "provider": provider},
		bson.M{
			"$set": bson.M{
				"external_user_id":  account.ExternalUserID,
				"external_username": account.Username,
				"access_token_enc":  accessEnc,
				"refresh_token_enc": refreshEnc,
				"scopes":            tokens.Scopes,
				"expires_at":        tokens.ExpiresAt,
				"status":            "active",
				"updated_at":        now,
			},
			"$setOnInsert": bson.M{"linked_at": now},
		},
		options.Update().SetUpsert(true))
	return err
}

// DisconnectMarketplace removes the user's linked account for a provider
func (h *Handlers) DisconnectMarketplace(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	p, ok := h.providerFromPath(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := h.db.Collection("linked_accounts").DeleteOne(ctx, bson.M{"user_id": userID, "provider": p.Name()})
	if err != nil {
		h.sendError(w, "Error disconnecting account", http.StatusInternalServerError, nil)
		return
	}
	if res.DeletedCount == 0 {
		h.sendError(w, "Account not linked", http.StatusNotFound, nil)
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"message": "Account disconnected"})
}

// accessTokenFor returns a usable access token, refreshing it if expired.
func (h *Handlers) accessTokenFor(ctx context.Context, userID primitive.ObjectID, p marketplace.Provider) (string, int, error) {
	var acct models.LinkedAccount
	err := h.db.Collection("linked_accounts").FindOne(ctx, bson.M{"user_id": userID, "provider": p.Name()}).Decode(&acct)
	if err == mongo.ErrNoDocuments || (err == nil && acct.Status == "revoked") {
		return "", http.StatusConflict, fmt.Errorf("link your %s account first", p.DisplayName())
	}
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("error loading linked account")
	}
	if time.Now().Before(acct.ExpiresAt) {
		token, err := h.cipher.Decrypt(acct.AccessTokenEnc)
		if err != nil {
			return "", http.StatusConflict, fmt.Errorf("stored credentials are invalid; please relink %s", p.DisplayName())
		}
		return token, 0, nil
	}

	refresh, err := h.cipher.Decrypt(acct.RefreshTokenEnc)
	if err != nil || refresh == "" {
		return "", http.StatusConflict, fmt.Errorf("%s session expired; please relink", p.DisplayName())
	}
	tokens, err := p.RefreshTokens(ctx, refresh)
	if err != nil {
		if _, uerr := h.db.Collection("linked_accounts").UpdateOne(ctx, bson.M{"_id": acct.ID}, bson.M{"$set": bson.M{"status": "expired"}}); uerr != nil {
			fmt.Printf("Warning: could not mark account expired: %v\n", uerr)
		}
		return "", http.StatusConflict, fmt.Errorf("%s session expired; please relink", p.DisplayName())
	}
	if err := h.storeLinkedAccount(ctx, userID, p.Name(), tokens, marketplace.Account{
		ExternalUserID: acct.ExternalUserID, Username: acct.ExternalUsername,
	}); err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("error saving refreshed credentials")
	}
	return tokens.AccessToken, 0, nil
}

// compsFor searches sold comparables for an owned item.
func (h *Handlers) compsFor(ctx context.Context, view *models.PortfolioItemView, source string) (models.CompsResponse, error) {
	q := marketplace.CompsQuery{
		CardID:    view.CardID,
		Keywords:  strings.TrimSpace(view.DisplayName + " " + view.DisplaySet),
		ItemType:  view.ItemType,
		Condition: view.Condition,
		Grading:   view.Grading,
		Source:    source,
		Days:      30,
	}
	if view.Card == nil {
		q.CardID = nil
	}
	sold, err := h.market.Comps().SearchSold(ctx, q)
	if err != nil {
		return models.CompsResponse{}, err
	}
	return models.CompsResponse{Query: q.Keywords, Sold: sold, Stats: marketplace.ComputeStats(sold)}, nil
}

// GetComps returns recent sold listings for an owned item
func (h *Handlers) GetComps(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	itemID, err := primitive.ObjectIDFromHex(r.URL.Query().Get("portfolio_item_id"))
	if err != nil {
		h.sendError(w, "Invalid portfolio_item_id", http.StatusBadRequest, nil)
		return
	}
	source := r.URL.Query().Get("provider")
	if source == "" {
		source = models.ProviderEbay
	}
	if _, ok := h.market.Provider(source); !ok {
		h.sendError(w, "Unknown marketplace provider", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	view, found, err := h.loadPortfolioItemView(ctx, userID, itemID)
	if err != nil {
		h.sendError(w, "Error retrieving item", http.StatusInternalServerError, nil)
		return
	}
	if !found {
		h.sendError(w, "Item not found", http.StatusNotFound, nil)
		return
	}
	comps, err := h.compsFor(ctx, view, source)
	if err != nil {
		h.sendError(w, "Error searching sold listings", http.StatusBadGateway, nil)
		return
	}
	sendJSON(w, http.StatusOK, comps)
}

// PrefillListing builds a listing draft for an owned item using sold comparables
func (h *Handlers) PrefillListing(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	var req models.PrefillRequest
	if err := decodeJSON(w, r, &req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest, nil)
		return
	}
	p, ok := h.market.Provider(req.Provider)
	if !ok {
		h.sendError(w, "Unknown marketplace provider", http.StatusBadRequest, nil)
		return
	}
	itemID, err := primitive.ObjectIDFromHex(req.PortfolioItemID)
	if err != nil {
		h.sendError(w, "Invalid portfolio_item_id", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	view, found, err := h.loadPortfolioItemView(ctx, userID, itemID)
	if err != nil {
		h.sendError(w, "Error retrieving item", http.StatusInternalServerError, nil)
		return
	}
	if !found {
		h.sendError(w, "Item not found", http.StatusNotFound, nil)
		return
	}
	comps, err := h.compsFor(ctx, view, p.Name())
	if err != nil {
		h.sendError(w, "Error searching sold listings", http.StatusBadGateway, nil)
		return
	}
	sendJSON(w, http.StatusOK, models.PrefillResponse{
		Draft: marketplace.BuildDraft(p.Name(), view, comps.Stats),
		Comps: comps,
	})
}

// validateListingDraft normalizes and validates a user-edited draft.
func validateListingDraft(d *models.ListingDraft, maxQty int) error {
	d.Title = cleanText(d.Title, false)
	d.Description = cleanText(d.Description, true)
	d.Condition = cleanText(d.Condition, false)
	d.CategoryID = cleanText(d.CategoryID, false)

	titleMax := marketplace.TCGPlayerTitleMax
	if d.Provider == models.ProviderEbay {
		titleMax = marketplace.EbayTitleMax
	}
	if d.Title == "" || len(d.Title) > titleMax {
		return fmt.Errorf("title must be 1-%d characters", titleMax)
	}
	if len([]rune(d.Description)) > 5000 {
		return fmt.Errorf("description must not exceed 5000 characters")
	}
	if d.Price < 0.01 || !validMoney(d.Price) {
		return fmt.Errorf("price must be at least $0.01")
	}
	if d.ShippingPrice < 0 || d.ShippingPrice > 1000 {
		return fmt.Errorf("shipping must be between 0 and 1000")
	}
	if d.Quantity < 1 || d.Quantity > maxQty {
		return fmt.Errorf("quantity must be between 1 and %d (quantity owned)", maxQty)
	}
	if d.Condition == "" || len(d.Condition) > 60 || len(d.CategoryID) > 60 {
		return fmt.Errorf("invalid condition or category")
	}
	if d.Format == "" {
		d.Format = "fixed_price"
	}
	if !validListingFormats[d.Format] {
		return fmt.Errorf("format must be fixed_price or auction")
	}
	if d.Provider == models.ProviderEbay {
		if d.Duration == "" {
			d.Duration = "GTC"
		}
		if !validEbayDurations[d.Duration] || (d.Format == "auction" && d.Duration == "GTC") {
			return fmt.Errorf("invalid listing duration")
		}
	} else {
		d.Duration = ""
		d.Format = "fixed_price"
	}
	if d.ImageURLs == nil {
		d.ImageURLs = []string{}
	}
	if len(d.ImageURLs) > 12 {
		return fmt.Errorf("at most 12 images")
	}
	for _, raw := range d.ImageURLs {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Host == "" || len(raw) > 500 {
			return fmt.Errorf("image URLs must be valid https URLs")
		}
	}
	if d.ItemSpecifics == nil {
		d.ItemSpecifics = []models.ItemSpecific{}
	}
	if len(d.ItemSpecifics) > 30 {
		return fmt.Errorf("at most 30 item specifics")
	}
	for i := range d.ItemSpecifics {
		s := &d.ItemSpecifics[i]
		s.Name, s.Value = cleanText(s.Name, false), cleanText(s.Value, false)
		if s.Name == "" || len(s.Name) > 65 || len(s.Value) > 100 {
			return fmt.Errorf("invalid item specific")
		}
	}
	return nil
}

// CreateListing saves a draft and optionally publishes it to the marketplace
func (h *Handlers) CreateListing(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	var req models.CreateListingRequest
	if err := decodeJSON(w, r, &req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest, nil)
		return
	}
	d := &req.Draft
	p, ok := h.market.Provider(d.Provider)
	if !ok {
		h.sendError(w, "Unknown marketplace provider", http.StatusBadRequest, nil)
		return
	}
	itemID, err := primitive.ObjectIDFromHex(d.PortfolioItemID)
	if err != nil {
		h.sendError(w, "Invalid portfolio_item_id", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	view, found, err := h.loadPortfolioItemView(ctx, userID, itemID)
	if err != nil {
		h.sendError(w, "Error retrieving item", http.StatusInternalServerError, nil)
		return
	}
	if !found {
		h.sendError(w, "Item not found", http.StatusNotFound, nil)
		return
	}
	if err := validateListingDraft(d, view.Quantity); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	token := ""
	if req.Publish {
		var status int
		token, status, err = h.accessTokenFor(ctx, userID, p)
		if err != nil {
			h.sendError(w, err.Error(), status, map[string]interface{}{"details": map[string]string{"code": "account_not_linked"}})
			return
		}
	}

	now := time.Now().UTC()
	listing := models.MarketplaceListing{
		UserID:          userID,
		PortfolioItemID: itemID,
		CardID:          view.CardID,
		Provider:        p.Name(),
		Status:          "draft",
		Title:           d.Title,
		Description:     d.Description,
		Price:           d.Price,
		Quantity:        d.Quantity,
		Condition:       d.Condition,
		CategoryID:      d.CategoryID,
		Format:          d.Format,
		Duration:        d.Duration,
		ShippingPrice:   d.ShippingPrice,
		ImageURLs:       d.ImageURLs,
		ItemSpecifics:   d.ItemSpecifics,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	status := http.StatusCreated
	if req.Publish {
		result, err := p.CreateListing(ctx, token, *d)
		if err != nil {
			listing.Status = "failed"
			listing.ErrorMessage = "Marketplace rejected the listing"
			status = http.StatusBadGateway
		} else {
			listing.Status = "published"
			listing.ExternalListingID = result.ExternalID
			listing.ListingURL = result.URL
			listing.PublishedAt = &now
		}
	}

	res, err := h.db.Collection("marketplace_listings").InsertOne(ctx, listing)
	if err != nil {
		h.sendError(w, "Error saving listing", http.StatusInternalServerError, nil)
		return
	}
	listing.ID = res.InsertedID.(primitive.ObjectID)

	if status == http.StatusBadGateway {
		h.sendError(w, listing.ErrorMessage, status, map[string]interface{}{"details": map[string]interface{}{"listing": listing}})
		return
	}
	sendJSON(w, status, listing)
}

// GetListings returns the user's marketplace listings, optionally for one item
func (h *Handlers) GetListings(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	filter := bson.M{"user_id": userID}
	if raw := r.URL.Query().Get("portfolio_item_id"); raw != "" {
		itemID, err := primitive.ObjectIDFromHex(raw)
		if err != nil {
			h.sendError(w, "Invalid portfolio_item_id", http.StatusBadRequest, nil)
			return
		}
		filter["portfolio_item_id"] = itemID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := h.db.Collection("marketplace_listings").Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(200))
	if err != nil {
		h.sendError(w, "Error retrieving listings", http.StatusInternalServerError, nil)
		return
	}
	defer cur.Close(ctx)
	listings := make([]models.MarketplaceListing, 0)
	if err := cur.All(ctx, &listings); err != nil {
		h.sendError(w, "Error decoding listings", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, listings)
}
