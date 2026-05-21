package client

import (
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "https://app.band.ai"
	defaultTimeout = 30 * time.Second
)

// Config controls the REST client behavior.
type Config struct {
	ApiKey     string
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
	Timeout    time.Duration
}

// Option configures a Client.
type Option func(*Config)

// WithApiKey configures the API key used for agent API authentication.
func WithApiKey(apiKey string) Option {
	return func(config *Config) {
		config.ApiKey = apiKey
	}
}

// WithBaseURL configures the base URL used for API requests.
func WithBaseURL(baseURL string) Option {
	return func(config *Config) {
		config.BaseURL = baseURL
	}
}

// WithHTTPClient configures the HTTP client used for requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(config *Config) {
		config.HTTPClient = httpClient
	}
}

// WithUserAgent configures the User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(config *Config) {
		config.UserAgent = userAgent
	}
}

// WithTimeout configures the timeout used by the default HTTP client.
func WithTimeout(timeout time.Duration) Option {
	return func(config *Config) {
		config.Timeout = timeout
	}
}
