package marketplace

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	mrand "math/rand"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/models"
	"github.com/jamesc159/monmetrics/internal/valuation"
)

// MockCodePrefix marks authorization codes issued by the mock consent route.
const MockCodePrefix = "mock_"

type mockProvider struct {
	name        string
	displayName string
	apiBaseURL  string
}

func newMockProvider(name, displayName, apiBaseURL string) *mockProvider {
	return &mockProvider{name: name, displayName: displayName, apiBaseURL: apiBaseURL}
}

func (m *mockProvider) Name() string        { return m.name }
func (m *mockProvider) DisplayName() string { return m.displayName }

func (m *mockProvider) AuthURL(state string) string {
	return fmt.Sprintf("%s/api/marketplace/%s/mock-authorize?state=%s", m.apiBaseURL, m.name, url.QueryEscape(state))
}

func (m *mockProvider) ExchangeCode(_ context.Context, code string) (Tokens, Account, error) {
	if !strings.HasPrefix(code, MockCodePrefix) || len(code) < len(MockCodePrefix)+8 {
		return Tokens{}, Account{}, errors.New("invalid authorization code")
	}
	sum := sha256.Sum256([]byte(m.name + code))
	id := hex.EncodeToString(sum[:])[:10]
	return Tokens{
			AccessToken:  "mock-access-" + randomHex(16),
			RefreshToken: "mock-refresh-" + randomHex(16),
			ExpiresAt:    time.Now().UTC().Add(2 * time.Hour),
			Scopes:       []string{"sell.inventory", "sell.account"},
		}, Account{
			ExternalUserID: m.name + "-" + id,
			Username:       "mock_" + m.name + "_seller_" + id[:4],
		}, nil
}

func (m *mockProvider) RefreshTokens(_ context.Context, refreshToken string) (Tokens, error) {
	if !strings.HasPrefix(refreshToken, "mock-refresh-") {
		return Tokens{}, errors.New("invalid refresh token")
	}
	return Tokens{
		AccessToken:  "mock-access-" + randomHex(16),
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().UTC().Add(2 * time.Hour),
		Scopes:       []string{"sell.inventory", "sell.account"},
	}, nil
}

func (m *mockProvider) CreateListing(_ context.Context, accessToken string, draft models.ListingDraft) (ListingResult, error) {
	if accessToken == "" {
		return ListingResult{}, errors.New("missing access token")
	}
	if draft.Price <= 0 || draft.Quantity <= 0 {
		return ListingResult{}, errors.New("price and quantity must be positive")
	}
	return ListingResult{ExternalID: strings.ToUpper(m.name) + "-MOCK-" + strings.ToUpper(randomHex(6))}, nil
}

// NewMockCode issues a mock OAuth authorization code.
func NewMockCode() string {
	return MockCodePrefix + randomHex(12)
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// mockComps derives "sold" listings from the card's stored price history so comps
// track the chart data, adjusted for the item's condition or grade.
type mockComps struct {
	db *mongo.Database
}

func (c *mockComps) SearchSold(ctx context.Context, q CompsQuery) ([]models.SoldListing, error) {
	sold := make([]models.SoldListing, 0)
	if q.CardID == nil {
		return sold, nil
	}
	days := q.Days
	if days <= 0 {
		days = 30
	}
	source := q.Source
	if source == "" {
		source = models.ProviderEbay
	}

	filter := bson.M{
		"card_id":   *q.CardID,
		"source":    source,
		"timestamp": bson.M{"$gte": time.Now().AddDate(0, 0, -days)},
	}
	cur, err := c.db.Collection("prices").Find(ctx, filter,
		options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(20))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var points []models.PricePoint
	if err := cur.All(ctx, &points); err != nil {
		return nil, err
	}

	mult, condLabel := compsAdjustment(q)
	h := fnv.New64a()
	h.Write([]byte(q.CardID.Hex() + q.Keywords))
	rng := mrand.New(mrand.NewSource(int64(h.Sum64())))

	for _, p := range points {
		price := p.Price * mult * (0.92 + rng.Float64()*0.16)
		shipping := 0.0
		if rng.Float64() < 0.5 {
			shipping = 4.99
		}
		sold = append(sold, models.SoldListing{
			Title:     strings.TrimSpace(q.Keywords + " " + condLabel),
			Price:     float64(int(price*100)) / 100,
			Shipping:  shipping,
			Condition: condLabel,
			SoldAt:    p.Timestamp.Add(time.Duration(rng.Intn(20)) * time.Hour),
			Source:    source,
		})
	}
	return sold, nil
}

func compsAdjustment(q CompsQuery) (float64, string) {
	switch q.ItemType {
	case models.ItemTypeGradedCard:
		if q.Grading != nil {
			return valuation.GradeMultiplier(q.Grading.Company, q.Grading.Grade),
				fmt.Sprintf("%s %s", q.Grading.Company, formatGrade(q.Grading.Grade))
		}
	case models.ItemTypeRawCard:
		if m, ok := valuation.ConditionMultipliers[q.Condition]; ok {
			return m, valuation.ConditionLabels[q.Condition]
		}
	case models.ItemTypeSealed:
		return 1, "Factory Sealed"
	}
	return 1, ""
}

func formatGrade(g float64) string {
	if g == float64(int(g)) {
		return fmt.Sprintf("%d", int(g))
	}
	return fmt.Sprintf("%.1f", g)
}
