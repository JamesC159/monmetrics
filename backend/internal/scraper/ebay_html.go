package scraper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
)

var ErrBlocked = errors.New("eBay blocked scraper request")

type EbayHTMLSource struct {
	BaseURL      string
	UserAgent    string
	RequestDelay time.Duration
	MaxPages     int
	Client       *http.Client
	throttleMu   sync.Mutex
	nextRequest  time.Time
}

func NewEbayHTMLSource(baseURL, userAgent string, requestDelay time.Duration, maxPages int) *EbayHTMLSource {
	return &EbayHTMLSource{
		BaseURL: strings.TrimRight(baseURL, "/"), UserAgent: userAgent,
		RequestDelay: requestDelay, MaxPages: maxPages,
		Client: &http.Client{Timeout: 25 * time.Second},
	}
}

func (s *EbayHTMLSource) Name() string { return models.ProviderEbay }

func (s *EbayHTMLSource) FetchSold(ctx context.Context, query SoldQuery) ([]models.SoldSale, error) {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 25 * time.Second}
	}
	baseURL := s.BaseURL
	if baseURL == "" {
		baseURL = "https://www.ebay.com"
	}
	maxPages := query.MaxPages
	if maxPages < 1 || (s.MaxPages > 0 && maxPages > s.MaxPages) {
		maxPages = s.MaxPages
	}
	if maxPages < 1 {
		maxPages = 3
	}

	all := make([]models.SoldSale, 0)
	for page := 1; page <= maxPages; page++ {
		if err := s.waitForRequest(ctx); err != nil {
			return nil, err
		}

		requestURL, err := buildSearchURL(baseURL, query.Keywords, query.Category, page)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", s.UserAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			return nil, fmt.Errorf("%w: HTTP %d", ErrBlocked, resp.StatusCode)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			return nil, fmt.Errorf("eBay search returned HTTP %d", resp.StatusCode)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		parsed, err := ParseSoldHTML(string(body), time.Now().UTC())
		if err != nil {
			return nil, err
		}
		if len(parsed) == 0 {
			break
		}
		oldest := time.Time{}
		for _, value := range parsed {
			if oldest.IsZero() || value.SoldAt.Before(oldest) {
				oldest = value.SoldAt
			}
			all = append(all, models.SoldSale{
				ItemID: value.ItemID, Source: s.Name(), Title: value.Title, Price: value.Price,
				Shipping: value.Shipping, Currency: value.Currency, SoldAt: value.SoldAt,
				BestOffer: value.BestOffer, URL: value.URL,
			})
		}
		if !query.StopAt.IsZero() && oldest.Before(query.StopAt) {
			break
		}
	}
	return all, nil
}

func (s *EbayHTMLSource) waitForRequest(ctx context.Context) error {
	s.throttleMu.Lock()
	now := time.Now()
	wait := s.nextRequest.Sub(now)
	delay := s.RequestDelay
	if delay > 0 {
		delay += time.Duration(rand.Int63n(int64(max(time.Second, delay/3))))
	}
	s.nextRequest = now.Add(max(wait, 0) + delay)
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

func buildSearchURL(baseURL, keywords, category string, page int) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/") + "/sch/i.html")
	if err != nil {
		return "", err
	}
	values := url.Values{
		"_nkw":        {keywords},
		"_sacat":      {category},
		"LH_Complete": {"1"},
		"LH_Sold":     {"1"},
		"_sop":        {"13"},
		"_ipg":        {"240"},
		"_pgn":        {strconv.Itoa(page)},
	}
	parsed.RawQuery = values.Encode()
	return parsed.String(), nil
}
