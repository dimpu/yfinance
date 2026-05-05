package yahoofinance

import (
	"net/http"
	"net/http/cookiejar"
	"sync"

	"golang.org/x/sync/semaphore"
)

const defaultQueryHost = "https://query2.finance.yahoo.com"
const defaultConcurrency = 4

type Client struct {
	httpClient   *http.Client
	queryHost    string
	sem          *semaphore.Weighted
	crumb        string
	crumbMu      sync.Mutex
	crumbValid   bool
	logger       Logger
	validation   ValidationOpts
	fetchOptions *FetchOptions
}

type Config struct {
	QueryHost    string
	Concurrency  int
	Logger       Logger
	HTTPClient   *http.Client
	Validation   ValidationOpts
	FetchOptions *FetchOptions
}

func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = &Config{}
	}
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = defaultConcurrency
	}
	queryHost := cfg.QueryHost
	if queryHost == "" {
		queryHost = defaultQueryHost
	}
	logger := cfg.Logger
	if logger == nil {
		logger = newDefaultLogger()
	}
	validation := cfg.Validation
	if !validation.LogErrors && !validation.AllowAdditionalProps {
		validation = ValidationOpts{LogErrors: true, AllowAdditionalProps: true}
	}
	jar, _ := cookiejar.New(nil)
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Jar: jar}
	}
	if httpClient.Jar == nil {
		httpClient.Jar = jar
	}
	return &Client{
		httpClient:   httpClient,
		queryHost:    queryHost,
		sem:          semaphore.NewWeighted(int64(concurrency)),
		logger:       logger,
		validation:   validation,
		fetchOptions: cfg.FetchOptions,
	}
}
