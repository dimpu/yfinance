package fetch

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	internalerrors "github.com/dimpu/yfinance/internal/errors"
	"github.com/dimpu/yfinance/internal/types"
)

func newTestFetcher(srv *httptest.Server) *Fetcher {
	jar, _ := cookiejar.New(nil)
	return NewFetcher(Config{
		QueryHost:   srv.URL,
		HTTPClient:  &http.Client{Jar: jar, Transport: srv.Client().Transport},
		Concurrency: 4,
		Logger:      types.NewDefaultLogger(),
	})
}

func TestFetchReturnsBodyOn200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/test" {
			t.Errorf("path = %s, want /test", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok-body")
	}))
	defer srv.Close()

	f := newTestFetcher(srv)
	body, err := f.Fetch(context.Background(), FetchConfig{URL: srv.URL + "/test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != "ok-body" {
		t.Errorf("body = %q, want %q", string(body), "ok-body")
	}
}

func TestFetchReturnsBadRequestErrorOn400(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "bad input")
	}))
	defer srv.Close()

	f := newTestFetcher(srv)
	_, err := f.Fetch(context.Background(), FetchConfig{URL: srv.URL + "/"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var bre *internalerrors.BadRequestError
	if !errors.As(err, &bre) {
		t.Errorf("error type = %T, want BadRequestError", err)
	}
}

func TestFetchReturnsHTTPErrorOnNon200(t *testing.T) {
	codes := []int{404, 500, 403, 429}
	for _, code := range codes {
		t.Run(fmt.Sprintf("%d", code), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
				fmt.Fprintf(w, "error-%d", code)
			}))
			defer srv.Close()

			f := newTestFetcher(srv)
			_, err := f.Fetch(context.Background(), FetchConfig{URL: srv.URL + "/"})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var he *internalerrors.HTTPError
			if !errors.As(err, &he) {
				t.Errorf("error type = %T, want HTTPError", err)
			}
			if he.StatusCode != code {
				t.Errorf("StatusCode = %d, want %d", he.StatusCode, code)
			}
		})
	}
}

func TestFetchSymbolSubstitution(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	f := newTestFetcher(srv)
	_, err := f.Fetch(context.Background(), FetchConfig{
		URL:    srv.URL + "/quote/{symbol}",
		Symbol: "AAPL",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/quote/AAPL"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestFetchQueryHostSubstitution(t *testing.T) {
	var gotHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	f := newTestFetcher(srv)
	_, err := f.Fetch(context.Background(), FetchConfig{
		URL: "${YF_QUERY_HOST}/v8/finance/chart/AAPL",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	srvURL, _ := url.Parse(srv.URL)
	if gotHost != srvURL.Host {
		t.Errorf("host = %q, want %q", gotHost, srvURL.Host)
	}
}

func TestFetchSemaphoreConcurrency(t *testing.T) {
	var (
		current    int64
		maxCurrent int64
		barrier    = make(chan struct{})
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt64(&current, 1)
		for {
			old := atomic.LoadInt64(&maxCurrent)
			if cur <= old || atomic.CompareAndSwapInt64(&maxCurrent, old, cur) {
				break
			}
		}
		// Wait until test signals us to proceed
		<-barrier
		atomic.AddInt64(&current, -1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	jar, _ := cookiejar.New(nil)
	f := NewFetcher(Config{
		QueryHost:   srv.URL,
		HTTPClient:  &http.Client{Jar: jar, Transport: srv.Client().Transport},
		Concurrency: 2,
		Logger:      types.NewDefaultLogger(),
	})

	// Launch 10 concurrent Fetch calls; semaphore limits to 2 at a time.
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.Fetch(context.Background(), FetchConfig{URL: srv.URL + "/"})
		}()
	}

	// Wait briefly for 2 requests to enter the handler, then observe max.
	time.Sleep(200 * time.Millisecond)
	mx := atomic.LoadInt64(&maxCurrent)

	// Unblock all handlers so goroutines can finish.
	close(barrier)
	wg.Wait()

	if mx > 2 {
		t.Errorf("max concurrent requests = %d, want <= 2", mx)
	}
	if mx < 1 {
		t.Errorf("max concurrent requests = %d, want >= 1", mx)
	}
}
