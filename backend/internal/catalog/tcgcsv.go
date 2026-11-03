package catalog

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ErrThrottled means the catalogue host asked us to slow down or blocked us.
var ErrThrottled = errors.New("catalog source throttled request")
var ErrNotFound = errors.New("catalog source resource not found")

const DefaultTCGCSVBaseURL = "https://tcgcsv.com"

// TCGCSVSource reads the public TCGplayer catalogue mirror at tcgcsv.com.
type TCGCSVSource struct {
	BaseURL      string
	UserAgent    string
	RequestDelay time.Duration
	Client       *http.Client

	throttleMu  sync.Mutex
	nextRequest time.Time
}

func NewTCGCSVSource(baseURL, userAgent string, requestDelay time.Duration) *TCGCSVSource {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultTCGCSVBaseURL
	}
	return &TCGCSVSource{
		BaseURL:      strings.TrimRight(baseURL, "/"),
		UserAgent:    userAgent,
		RequestDelay: requestDelay,
		Client:       &http.Client{Timeout: 45 * time.Second},
	}
}

func (s *TCGCSVSource) Name() string { return "tcgcsv" }

type envelope[T any] struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors"`
	Results []T      `json:"results"`
}

func (s *TCGCSVSource) Categories(ctx context.Context) ([]Category, error) {
	return fetchList[Category](ctx, s, "/tcgplayer/categories")
}

func (s *TCGCSVSource) Groups(ctx context.Context, categoryID int) ([]Group, error) {
	return fetchList[Group](ctx, s, fmt.Sprintf("/tcgplayer/%d/groups", categoryID))
}

func (s *TCGCSVSource) Products(ctx context.Context, categoryID, groupID int) ([]Product, error) {
	return fetchList[Product](ctx, s, fmt.Sprintf("/tcgplayer/%d/%d/products", categoryID, groupID))
}

// Prices returns the current market prices for a group. One product yields one row per
// printing (subTypeName), so results are many-to-one against products.
func (s *TCGCSVSource) Prices(ctx context.Context, categoryID, groupID int) ([]ProductPrice, error) {
	return fetchList[ProductPrice](ctx, s, fmt.Sprintf("/tcgplayer/%d/%d/prices", categoryID, groupID))
}

// ProductsAndPrices reads a group's products and prices from its one CSV file, half the
// requests of fetching the products and prices JSON separately.
func (s *TCGCSVSource) ProductsAndPrices(ctx context.Context, categoryID, groupID int) ([]Product, []ProductPrice, error) {
	path := fmt.Sprintf("/tcgplayer/%d/%d/ProductsAndPrices.csv", categoryID, groupID)
	body, err := s.get(ctx, path)
	if err != nil {
		return nil, nil, err
	}
	products, prices, err := parseProductsAndPrices(bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return products, prices, nil
}

// tcgplayerProductURL replaces the CSV's url column, which is a TCGCSV short link rather
// than the product page. TCGplayer resolves the id without the slug.
const tcgplayerProductURL = "https://www.tcgplayer.com/product/%d"

// parseProductsAndPrices reads ProductsAndPrices.csv: one row per product and printing,
// with extendedData flattened into ext<Name> columns whose set varies by game. A product
// with no prices yet has a single row with an empty subTypeName.
func parseProductsAndPrices(r io.Reader) ([]Product, []ProductPrice, error) {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	column := make(map[string]int, len(header))
	var extColumns []string
	for i, name := range header {
		column[name] = i
		if strings.HasPrefix(name, "ext") {
			extColumns = append(extColumns, name)
		}
	}
	for _, name := range []string{"productId", "name", "subTypeName"} {
		if _, ok := column[name]; !ok {
			return nil, nil, fmt.Errorf("missing %s column", name)
		}
	}

	var products []Product
	var prices []ProductPrice
	seen := map[int]bool{}
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		field := func(name string) string {
			if i, ok := column[name]; ok {
				return strings.TrimSpace(record[i])
			}
			return ""
		}
		line, _ := reader.FieldPos(0)
		productID, err := strconv.Atoi(field("productId"))
		if err != nil {
			return nil, nil, fmt.Errorf("line %d: productId: %w", line, err)
		}

		if !seen[productID] {
			seen[productID] = true
			categoryID, _ := strconv.Atoi(field("categoryId"))
			groupID, _ := strconv.Atoi(field("groupId"))
			product := Product{
				ProductID:  productID,
				CategoryID: categoryID,
				GroupID:    groupID,
				Name:       field("name"),
				CleanName:  field("cleanName"),
				ImageURL:   field("imageUrl"),
				URL:        fmt.Sprintf(tcgplayerProductURL, productID),
				ModifiedOn: field("modifiedOn"),
			}
			for _, name := range extColumns {
				if value := field(name); value != "" {
					product.ExtendedData = append(product.ExtendedData,
						ExtendedField{Name: strings.TrimPrefix(name, "ext"), Value: value})
				}
			}
			products = append(products, product)
		}

		subType := field("subTypeName")
		if subType == "" {
			continue
		}
		price := ProductPrice{ProductID: productID, SubTypeName: subType}
		for _, target := range []struct {
			name string
			dest **float64
		}{
			{"lowPrice", &price.LowPrice},
			{"midPrice", &price.MidPrice},
			{"highPrice", &price.HighPrice},
			{"marketPrice", &price.MarketPrice},
			{"directLowPrice", &price.DirectLowPrice},
		} {
			raw := field(target.name)
			if raw == "" {
				continue // null in the JSON feed
			}
			value, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("line %d: %s: %w", line, target.name, err)
			}
			*target.dest = &value
		}
		prices = append(prices, price)
	}
	return products, prices, nil
}

// LastUpdated reports when the mirror last refreshed, so cycles can be skipped.
func (s *TCGCSVSource) LastUpdated(ctx context.Context) (time.Time, error) {
	body, err := s.get(ctx, "/last-updated.txt")
	if err != nil {
		return time.Time{}, err
	}
	stamp, err := time.Parse(time.RFC3339, strings.TrimSpace(string(body)))
	if err != nil {
		return time.Time{}, fmt.Errorf("parse last-updated: %w", err)
	}
	return stamp, nil
}

func fetchList[T any](ctx context.Context, s *TCGCSVSource, path string) ([]T, error) {
	body, err := s.get(ctx, path)
	if err != nil {
		return nil, err
	}
	var payload envelope[T]
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if !payload.Success && len(payload.Errors) > 0 {
		return nil, fmt.Errorf("catalog source %s: %s", path, strings.Join(payload.Errors, "; "))
	}
	return payload.Results, nil
}

func (s *TCGCSVSource) get(ctx context.Context, path string) ([]byte, error) {
	if err := s.waitForRequest(ctx); err != nil {
		return nil, err
	}
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	base := s.BaseURL
	if base == "" {
		base = DefaultTCGCSVBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", s.UserAgent)
	req.Header.Set("Accept", "application/json,text/csv,text/plain")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("%w: HTTP %d", ErrThrottled, resp.StatusCode)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: catalog source %s returned HTTP %d", ErrNotFound, path, resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("catalog source %s returned HTTP %d", path, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 64<<20))
}

func (s *TCGCSVSource) waitForRequest(ctx context.Context) error {
	s.throttleMu.Lock()
	now := time.Now()
	wait := s.nextRequest.Sub(now)
	delay := s.RequestDelay
	if delay > 0 {
		delay += time.Duration(rand.Int63n(int64(delay/2 + time.Millisecond)))
	}
	if wait < 0 {
		wait = 0
	}
	s.nextRequest = now.Add(wait + delay)
	s.throttleMu.Unlock()
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
