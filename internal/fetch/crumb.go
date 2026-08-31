package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	errpkg "github.com/dimpu/yfinance/internal/errors"
)

const (
	maxRetries     = 5
	baseBackoff    = 5 * time.Second
	maxBackoff     = 60 * time.Second
)

func backoff(attempt int) time.Duration {
	d := time.Duration(math.Pow(2, float64(attempt))) * baseBackoff
	if d > maxBackoff {
		return maxBackoff
	}
	return d
}

func retryAfter(resp *http.Response) time.Duration {
	v := resp.Header.Get("Retry-After")
	if v == "" {
		return 0
	}
	if d, err := strconv.Atoi(v); err == nil {
		return time.Duration(d) * time.Second
	}
	return 0
}

func (f *Fetcher) ensureCrumb(ctx context.Context) error {
	f.crumbMu.Lock()
	defer f.crumbMu.Unlock()

	if f.crumbValid && f.crumb != "" {
		return nil
	}

	if err := f.fetchCookies(ctx); err != nil {
		return fmt.Errorf("fetching cookies: %w", err)
	}

	crumb, err := f.fetchCrumb(ctx)
	if err != nil {
		return fmt.Errorf("fetching crumb: %w", err)
	}

	f.crumb = crumb
	f.crumbValid = true
	return nil
}

func (f *Fetcher) fetchCookies(ctx context.Context) error {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			wait := backoff(attempt - 1)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		resp, err := f.doFetchCookies(ctx)
		if err == nil {
			return nil
		}
		if !isRetryable(err) {
			return err
		}
		lastErr = err
		f.logger.Warn("fetchCookies attempt %d failed: %v, retrying in %v...", attempt+1, err, backoff(attempt))
		_ = resp
	}
	return fmt.Errorf("fetching cookies after %d retries: %w", maxRetries, lastErr)
}

func (f *Fetcher) doFetchCookies(ctx context.Context) (*http.Response, error) {
	reqURL := "https://finance.yahoo.com/quote/AAPL"
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
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusTooManyRequests {
		return resp, &errpkg.HTTPError{StatusCode: resp.StatusCode, Body: "too many requests"}
	}

	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		loc := resp.Header.Get("Location")
		if loc != "" {
			return nil, f.handleConsentRedirect(ctx, loc)
		}
	}
	return nil, nil
}

func (f *Fetcher) handleConsentRedirect(ctx context.Context, redirectURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, redirectURL, nil)
	if err != nil {
		return err
	}
	f.setHeaders(req)
	client := &http.Client{
		Jar:       f.httpClient.Jar,
		Transport: f.httpClient.Transport,
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

	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		loc := resp.Header.Get("Location")
		if loc != "" && !strings.Contains(loc, "guce.yahoo.com") {
			return f.handleConsentRedirect(ctx, loc)
		}
	}
	return nil
}

func (f *Fetcher) fetchCrumb(ctx context.Context) (string, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			wait := backoff(attempt - 1)
			f.logger.Warn("fetchCrumb attempt %d failed: %v, retrying in %v...", attempt, lastErr, wait)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(wait):
			}
		}

		crumb, resp, err := f.doFetchCrumb(ctx)
		if err == nil {
			return crumb, nil
		}
		if !isRetryable(err) {
			return "", err
		}
		if ra := retryAfter(resp); ra > 0 {
			f.logger.Warn("fetchCrumb server requested Retry-After: %v", ra)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(ra):
			}
			crumb, _, retryErr := f.doFetchCrumb(ctx)
			if retryErr == nil {
				return crumb, nil
			}
			err = retryErr
		}
		lastErr = err
	}
	return "", fmt.Errorf("fetching crumb after %d retries: %w", maxRetries, lastErr)
}

func (f *Fetcher) doFetchCrumb(ctx context.Context) (string, *http.Response, error) {
	reqURL := "https://query1.finance.yahoo.com/v1/test/getcrumb"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", nil, err
	}
	f.setHeaders(req)
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", resp, &errpkg.HTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	return string(body), resp, nil
}

func (f *Fetcher) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "yahoo-finance-go/1.0")
	if f.fetchOptions != nil {
		for k, v := range f.fetchOptions.Headers {
			req.Header.Set(k, v)
		}
	}
}

func (f *Fetcher) invalidateCrumb() {
	f.crumbMu.Lock()
	defer f.crumbMu.Unlock()
	f.crumb = ""
	f.crumbValid = false
}

func isRetryable(err error) bool {
	var httpErr *errpkg.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == http.StatusTooManyRequests ||
			httpErr.StatusCode == http.StatusServiceUnavailable ||
			httpErr.StatusCode == http.StatusBadGateway ||
			httpErr.StatusCode == http.StatusGatewayTimeout
	}
	return false
}

func (f *Fetcher) fetchWithCrumb(ctx context.Context, reqURL string) (*http.Response, error) {
	if err := f.ensureCrumb(ctx); err != nil {
		return nil, err
	}
	u, err := url.Parse(reqURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("crumb", f.crumb)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	f.setHeaders(req)
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		f.invalidateCrumb()
		if err := f.ensureCrumb(ctx); err != nil {
			return nil, err
		}
		q.Set("crumb", f.crumb)
		u.RawQuery = q.Encode()
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		f.setHeaders(req)
		resp, err = f.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}
