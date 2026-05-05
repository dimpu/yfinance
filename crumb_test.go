package yahoofinance

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"golang.org/x/sync/semaphore"
)

// redirectTransport routes requests to the given httptest.Server regardless
// of the original URL host. This lets us test code with hardcoded URLs.
type redirectTransport struct {
	server *httptest.Server
	base   http.RoundTripper
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point at the test server while preserving path and query.
	newURL := *req.URL
	newURL.Scheme = "http"
	newURL.Host = t.server.Listener.Addr().String()
	req.URL = &newURL
	return t.base.RoundTrip(req)
}

// newTestClient creates a Client whose HTTP requests are routed to srv.
func newTestClient(srv *httptest.Server) *Client {
	jar, _ := cookiejar.New(nil)
	transport := &redirectTransport{
		server: srv,
		base:   http.DefaultTransport,
	}
	return &Client{
		httpClient: &http.Client{
			Jar:       jar,
			Transport: transport,
		},
		sem:          semaphore.NewWeighted(4),
		logger:     newDefaultLogger(),
		validation: ValidationOpts{LogErrors: true, AllowAdditionalProps: true},
	}
}

// --- Tests ---

func TestEnsureCrumb_FetchesCookiesThenCrumb(t *testing.T) {
	var cookieCalls int32
	var crumbCalls int32

	mux := http.NewServeMux()
	// Endpoint hit by fetchCookies
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&cookieCalls, 1)
		w.WriteHeader(http.StatusOK)
	})
	// Endpoint hit by fetchCrumb
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&crumbCalls, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testcrumb123"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	err := client.ensureCrumb(context.Background())
	if err != nil {
		t.Fatalf("ensureCrumb returned error: %v", err)
	}

	if atomic.LoadInt32(&cookieCalls) != 1 {
		t.Errorf("expected cookie endpoint called once, got %d", cookieCalls)
	}
	if atomic.LoadInt32(&crumbCalls) != 1 {
		t.Errorf("expected crumb endpoint called once, got %d", crumbCalls)
	}
	if client.crumb != "testcrumb123" {
		t.Errorf("expected crumb 'testcrumb123', got %q", client.crumb)
	}
	if !client.crumbValid {
		t.Error("expected crumbValid to be true")
	}
}

func TestEnsureCrumb_CachedOnSecondCall(t *testing.T) {
	var crumbCalls int32

	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&crumbCalls, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("cachedcrumb"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	// First call: should fetch crumb
	if err := client.ensureCrumb(context.Background()); err != nil {
		t.Fatalf("first ensureCrumb: %v", err)
	}
	// Second call: should use cached crumb
	if err := client.ensureCrumb(context.Background()); err != nil {
		t.Fatalf("second ensureCrumb: %v", err)
	}

	if atomic.LoadInt32(&crumbCalls) != 1 {
		t.Errorf("expected crumb endpoint called once (cached on second call), got %d", crumbCalls)
	}
}

func TestFetchWithCrumb_AppendsCrumbParam(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mycrumb"))
	})

	var dataRequestCrumb string
	mux.HandleFunc("/v8/finance/chart/AAPL", func(w http.ResponseWriter, r *http.Request) {
		dataRequestCrumb = r.URL.Query().Get("crumb")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	resp, err := client.fetchWithCrumb(context.Background(), "https://query2.finance.yahoo.com/v8/finance/chart/AAPL")
	if err != nil {
		t.Fatalf("fetchWithCrumb error: %v", err)
	}
	defer resp.Body.Close()

	if dataRequestCrumb != "mycrumb" {
		t.Errorf("expected crumb query param 'mycrumb', got %q", dataRequestCrumb)
	}
}

func TestFetchWithCrumb_InvalidatesOn401(t *testing.T) {
	var crumbCalls int32
	var dataCalls int32

	mux := http.NewServeMux()
	mux.HandleFunc("/quote/AAPL", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/test/getcrumb", func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&crumbCalls, 1)
		if count == 1 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("oldcrumb"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("newcrumb"))
		}
	})

	mux.HandleFunc("/v8/finance/chart/AAPL", func(w http.ResponseWriter, r *http.Request) {
		call := atomic.AddInt32(&dataCalls, 1)
		if call == 1 {
			// First request: reject with 401 to trigger crumb invalidation
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized"))
			return
		}
		// Second request: succeed
		crumb := r.URL.Query().Get("crumb")
		if crumb != "newcrumb" {
			t.Errorf("retry expected crumb 'newcrumb', got %q", crumb)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newTestClient(srv)

	resp, err := client.fetchWithCrumb(context.Background(), "https://query2.finance.yahoo.com/v8/finance/chart/AAPL")
	if err != nil {
		t.Fatalf("fetchWithCrumb error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&crumbCalls) != 2 {
		t.Errorf("expected crumb endpoint called twice (initial + refetch after 401), got %d", crumbCalls)
	}
	if atomic.LoadInt32(&dataCalls) != 2 {
		t.Errorf("expected data endpoint called twice (initial 401 + retry), got %d", dataCalls)
	}
	if client.crumb != "newcrumb" {
		t.Errorf("expected crumb 'newcrumb' after retry, got %q", client.crumb)
	}
}
