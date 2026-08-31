package yahoofinance

import (
	"net/http"
	"net/http/cookiejar"

	"github.com/dimpu/yfinance/internal/types"
)

const (
	defaultQueryHost   = "https://query2.finance.yahoo.com"
	defaultConcurrency = 4
)

// Config holds configuration for creating a Client.
type Config struct {
	QueryHost    string
	Concurrency  int
	Logger       types.Logger
	HTTPClient   *http.Client
	Validation   types.ValidationOpts
	FetchOptions *types.FetchOptions
}

func applyDefaults(cfg *Config) *Config {
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
		logger = types.NewDefaultLogger()
	}
	validation := cfg.Validation
	if !validation.LogErrors && !validation.AllowAdditionalProps {
		validation = types.ValidationOpts{LogErrors: true, AllowAdditionalProps: true}
	}
	jar, _ := cookiejar.New(nil)
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Jar: jar}
	}
	if httpClient.Jar == nil {
		httpClient.Jar = jar
	}

	cfg.QueryHost = queryHost
	cfg.Concurrency = concurrency
	cfg.Logger = logger
	cfg.Validation = validation
	cfg.HTTPClient = httpClient
	return cfg
}
