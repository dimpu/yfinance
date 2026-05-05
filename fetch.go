package yahoofinance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type fetchConfig struct {
	url        string
	needsCrumb bool
	symbol     string // for URL path substitution
}

func (c *Client) fetch(ctx context.Context, cfg fetchConfig, opts *ModuleOptions) ([]byte, error) {
	// Acquire semaphore slot
	if err := c.sem.Acquire(ctx, 1); err != nil {
		return nil, fmt.Errorf("acquiring semaphore: %w", err)
	}
	defer c.sem.Release(1)

	// Build URL
	reqURL := cfg.url
	reqURL = strings.ReplaceAll(reqURL, "${YF_QUERY_HOST}", c.queryHost)
	if cfg.symbol != "" {
		reqURL = strings.ReplaceAll(reqURL, "{symbol}", url.PathEscape(cfg.symbol))
	}

	var resp *http.Response
	var err error

	if cfg.needsCrumb {
		resp, err = c.fetchWithCrumb(ctx, reqURL)
	} else {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		c.setHeaders(req)
		resp, err = c.httpClient.Do(req)
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
		return nil, &BadRequestError{Message: string(body)}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
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
			return &BadRequestError{Message: errResp.Finance.Error.Description}
		}
		return &HTTPError{StatusCode: 200, Body: errResp.Finance.Error.Description}
	}
	return nil
}
