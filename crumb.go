package yahoofinance

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) ensureCrumb(ctx context.Context) error {
	c.crumbMu.Lock()
	defer c.crumbMu.Unlock()

	if c.crumbValid && c.crumb != "" {
		return nil
	}

	// Step 1: Visit finance.yahoo.com to get initial cookies
	if err := c.fetchCookies(ctx); err != nil {
		return fmt.Errorf("fetching cookies: %w", err)
	}

	// Step 2: Fetch crumb token
	crumb, err := c.fetchCrumb(ctx)
	if err != nil {
		return fmt.Errorf("fetching crumb: %w", err)
	}

	c.crumb = crumb
	c.crumbValid = true
	return nil
}

func (c *Client) fetchCookies(ctx context.Context) error {
	reqURL := "https://finance.yahoo.com/quote/AAPL"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // drain body

	// Handle GDPR consent redirect chain
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		loc := resp.Header.Get("Location")
		if loc != "" {
			return c.handleConsentRedirect(ctx, loc)
		}
	}
	return nil
}

func (c *Client) handleConsentRedirect(ctx context.Context, redirectURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, redirectURL, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	// Don't follow redirects automatically - we need to collect cookies
	client := &http.Client{
		Jar:       c.httpClient.Jar,
		Transport: c.httpClient.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	// Follow the consent chain
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		loc := resp.Header.Get("Location")
		if loc != "" && !strings.Contains(loc, "guce.yahoo.com") {
			return c.handleConsentRedirect(ctx, loc)
		}
	}
	return nil
}

func (c *Client) fetchCrumb(ctx context.Context) (string, error) {
	reqURL := "https://query1.finance.yahoo.com/v1/test/getcrumb"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", &HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	return string(body), nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "yahoo-finance-go/1.0")
	if c.fetchOptions != nil {
		for k, v := range c.fetchOptions.Headers {
			req.Header.Set(k, v)
		}
	}
}

func (c *Client) invalidateCrumb() {
	c.crumbMu.Lock()
	defer c.crumbMu.Unlock()
	c.crumb = ""
	c.crumbValid = false
}

func (c *Client) fetchWithCrumb(ctx context.Context, reqURL string) (*http.Response, error) {
	if err := c.ensureCrumb(ctx); err != nil {
		return nil, err
	}
	u, err := url.Parse(reqURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("crumb", c.crumb)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	// On 401, invalidate crumb and retry once
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.invalidateCrumb()
		if err := c.ensureCrumb(ctx); err != nil {
			return nil, err
		}
		q.Set("crumb", c.crumb)
		u.RawQuery = q.Encode()
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		c.setHeaders(req)
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}
