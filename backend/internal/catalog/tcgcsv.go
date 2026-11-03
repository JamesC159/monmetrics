package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
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
	req.Header.Set("Accept", "application/json,text/plain")
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
