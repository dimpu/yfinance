package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/types"
	"golang.org/x/sync/semaphore"
)

// Config holds configuration for creating a Fetcher.
type Config struct {
	QueryHost    string
	HTTPClient   *http.Client
	Concurrency  int
	Logger       types.Logger
	FetchOptions *types.FetchOptions
}

// FetchConfig controls a single fetch request.
type FetchConfig struct {
	URL        string
	NeedsCrumb bool
	Symbol     string
}

// Fetcher handles HTTP requests to Yahoo Finance APIs.
type Fetcher struct {
	httpClient   *http.Client
	queryHost    string
	sem          *semaphore.Weighted
	logger       types.Logger
	fetchOptions *types.FetchOptions

	crumb      string
	crumbMu    sync.Mutex
	crumbValid bool
}

// NewFetcher creates a new Fetcher with the given configuration.
func NewFetcher(cfg Config) *Fetcher {
	return &Fetcher{
		httpClient:   cfg.HTTPClient,
		queryHost:    cfg.QueryHost,
		sem:          semaphore.NewWeighted(int64(cfg.Concurrency)),
		logger:       cfg.Logger,
		fetchOptions: cfg.FetchOptions,
	}
}

// Fetch performs an HTTP GET request with optional crumb authentication.
func (f *Fetcher) Fetch(ctx context.Context, cfg FetchConfig) ([]byte, error) {
	if err := f.sem.Acquire(ctx, 1); err != nil {
		return nil, fmt.Errorf("acquiring semaphore: %w", err)
	}
	defer f.sem.Release(1)

	reqURL := cfg.URL
	reqURL = strings.ReplaceAll(reqURL, "${YF_QUERY_HOST}", f.queryHost)
	if cfg.Symbol != "" {
		reqURL = strings.ReplaceAll(reqURL, "{symbol}", url.PathEscape(cfg.Symbol))
	}

	var resp *http.Response
	var err error

	if cfg.NeedsCrumb {
		resp, err = f.fetchWithCrumb(ctx, reqURL)
	} else {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		f.setHeaders(req)
		resp, err = f.httpClient.Do(req)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return nil, &errors.BadRequestError{Message: string(body)}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &errors.HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	return body, nil
}

// Do performs a direct HTTP GET request without semaphore or crumb logic.
func (f *Fetcher) Do(ctx context.Context, reqURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	f.setHeaders(req)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return nil, &errors.BadRequestError{Message: string(body)}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &errors.HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	return body, nil
}

// parseYahooError extracts Yahoo's error format from response body.
func parseYahooError(body []byte) error {
	var errResp struct {
		Finance struct {
			Error struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"error"`
		} `json:"finance"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return nil
	}
	if errResp.Finance.Error.Code != "" {
		if errResp.Finance.Error.Code == "Bad Request" {
			return &errors.BadRequestError{Message: errResp.Finance.Error.Description}
		}
		return &errors.HTTPError{StatusCode: 200, Body: errResp.Finance.Error.Description}
	}
	return nil
}
