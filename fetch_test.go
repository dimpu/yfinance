package yahoofinance

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchReturnsBodyOn200(t *testing.T) {
	wantBody := `{"result":"ok"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(wantBody))
	}))
	defer srv.Close()

	client := NewClient(&Config{QueryHost: "test"})
	// Override httpClient to point at test server
	client.httpClient = srv.Client()

	// We need to redirect the URL to the test server.
	// Use a fetchConfig that does not need crumb, with URL pointing at test server.
	cfg := fetchConfig{
		url:        srv.URL + "/test",
		needsCrumb: false,
	}

	body, err := client.fetch(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(body) != wantBody {
		t.Errorf("expected body %q, got %q", wantBody, string(body))
	}
}

func TestFetchReturnsBadRequestErrorOn400(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request data"))
	}))
	defer srv.Close()

	client := NewClient(&Config{QueryHost: "test"})
	client.httpClient = srv.Client()

	cfg := fetchConfig{
		url:        srv.URL + "/test",
		needsCrumb: false,
	}

	_, err := client.fetch(context.Background(), cfg, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var badReqErr *BadRequestError
	if !errors.As(err, &badReqErr) {
		t.Errorf("expected BadRequestError, got %T: %v", err, err)
	}
	if badReqErr.Message != "invalid request data" {
		t.Errorf("expected message 'invalid request data', got %q", badReqErr.Message)
	}
}

func TestFetchReturnsHTTPErrorOnNon200Non400(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"404 Not Found", http.StatusNotFound, "not found"},
		{"500 Internal Server Error", http.StatusInternalServerError, "server error"},
		{"403 Forbidden", http.StatusForbidden, "forbidden"},
		{"429 Too Many Requests", http.StatusTooManyRequests, "rate limited"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			client := NewClient(&Config{QueryHost: "test"})
			client.httpClient = srv.Client()

			cfg := fetchConfig{
				url:        srv.URL + "/test",
				needsCrumb: false,
			}

			_, err := client.fetch(context.Background(), cfg, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var httpErr *HTTPError
			if !errors.As(err, &httpErr) {
				t.Errorf("expected HTTPError, got %T: %v", err, err)
			}
			if httpErr.StatusCode != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, httpErr.StatusCode)
			}
			if httpErr.Body != tt.body {
				t.Errorf("expected body %q, got %q", tt.body, httpErr.Body)
			}
		})
	}
}

func TestParseYahooError(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		wantErrType interface{} // *BadRequestError or *HTTPError
		wantMsg     string
	}{
		{
			name:        "Bad Request code returns BadRequestError",
			body:        `{"finance":{"error":{"code":"Bad Request","description":"Invalid symbol"}}}`,
			wantErrType: &BadRequestError{},
			wantMsg:     "Invalid symbol",
		},
		{
			name:        "Other error code returns HTTPError",
			body:        `{"finance":{"error":{"code":"Not Found","description":"No data found"}}}`,
			wantErrType: &HTTPError{},
			wantMsg:     "No data found",
		},
		{
			name:        "No error code returns nil",
			body:        `{"finance":{"error":{"code":"","description":""}}}`,
			wantErrType: nil,
			wantMsg:     "",
		},
		{
			name:        "Invalid JSON returns nil",
			body:        `not json at all`,
			wantErrType: nil,
			wantMsg:     "",
		},
		{
			name:        "Empty object returns nil",
			body:        `{}`,
			wantErrType: nil,
			wantMsg:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseYahooError([]byte(tt.body))
			if tt.wantErrType == nil {
				if err != nil {
					t.Errorf("expected nil, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			switch tt.wantErrType.(type) {
			case *BadRequestError:
				var badReqErr *BadRequestError
				if !errors.As(err, &badReqErr) {
					t.Errorf("expected BadRequestError, got %T: %v", err, err)
				}
				if badReqErr.Message != tt.wantMsg {
					t.Errorf("expected message %q, got %q", tt.wantMsg, badReqErr.Message)
				}
			case *HTTPError:
				var httpErr *HTTPError
				if !errors.As(err, &httpErr) {
					t.Errorf("expected HTTPError, got %T: %v", err, err)
				}
				if httpErr.Body != tt.wantMsg {
					t.Errorf("expected body %q, got %q", tt.wantMsg, httpErr.Body)
				}
			}
		})
	}
}

func TestFetchSemaphoreConcurrency(t *testing.T) {
	const maxConcurrent = 2
	var currentRequests int64
	var maxObserved int64
	var mu sync.Mutex

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt64(&currentRequests, 1)
		defer atomic.AddInt64(&currentRequests, -1)

		mu.Lock()
		if current > maxObserved {
			maxObserved = current
		}
		mu.Unlock()

		// Hold the request open briefly to increase concurrency overlap
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	client := NewClient(&Config{
		QueryHost:   "test",
		Concurrency: maxConcurrent,
	})
	client.httpClient = srv.Client()

	const totalRequests = 10
	var wg sync.WaitGroup
	wg.Add(totalRequests)

	for i := 0; i < totalRequests; i++ {
		go func() {
			defer wg.Done()
			cfg := fetchConfig{
				url:        srv.URL + "/test",
				needsCrumb: false,
			}
			_, err := client.fetch(context.Background(), cfg, nil)
			if err != nil {
				t.Errorf("fetch failed: %v", err)
			}
		}()
	}
	wg.Wait()

	mu.Lock()
	observed := maxObserved
	mu.Unlock()

	if observed > maxConcurrent {
		t.Errorf("expected max concurrent requests <= %d, got %d", maxConcurrent, observed)
	}
}

func TestFetchSymbolSubstitution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the symbol is properly escaped in the path
		if r.URL.Path != "/quote/AAPL" {
			t.Errorf("expected path /quote/AAPL, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	client := NewClient(&Config{QueryHost: "test"})
	client.httpClient = srv.Client()

	cfg := fetchConfig{
		url:        srv.URL + "/quote/{symbol}",
		needsCrumb: false,
		symbol:     "AAPL",
	}

	_, err := client.fetch(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFetchQueryHostSubstitution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	client := NewClient(&Config{QueryHost: srv.URL})
	// We need the httpClient to use the test server's client, but also
	// have the URL substitution work. Since the URL will have ${YF_QUERY_HOST}
	// replaced with srv.URL, and the client uses srv.Client(), this should work.
	// However, the default httpClient already has a cookie jar from NewClient.
	// We need to keep that jar but use the test server's transport.
	srvClient := srv.Client()
	client.httpClient.Transport = srvClient.Transport

	cfg := fetchConfig{
		url:        "${YF_QUERY_HOST}/v8/finance/chart/AAPL",
		needsCrumb: false,
	}

	_, err := client.fetch(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
