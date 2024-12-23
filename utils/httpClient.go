package utils

import (
	"crypto/tls"
	"net/http"
	"time"
)

const (
	// DefaultTimeout is the default timeout for HTTP requests
	DefaultTimeout = 30 * time.Second
	// DefaultMaxIdleConns is the default maximum number of idle connections
	DefaultMaxIdleConns = 100
	// DefaultIdleConnTimeout is the default timeout for idle connections
	DefaultIdleConnTimeout = 90 * time.Second
)

// HTTPClientConfig holds configuration for the HTTP client
type HTTPClientConfig struct {
	// Timeout is the maximum time to wait for a response
	Timeout time.Duration
	// SkipTLSVerification disables TLS certificate verification
	SkipTLSVerification bool
	// MaxIdleConns controls the maximum number of idle connections
	MaxIdleConns int
	// IdleConnTimeout is the maximum amount of time an idle connection will be kept in the pool
	IdleConnTimeout time.Duration
}

// NewDefaultConfig returns a new HTTPClientConfig with default values
func NewDefaultConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		Timeout:             DefaultTimeout,
		SkipTLSVerification: false,
		MaxIdleConns:        DefaultMaxIdleConns,
		IdleConnTimeout:     DefaultIdleConnTimeout,
	}
}

// CreateHTTPClient creates an HTTP client with the specified configuration
func CreateHTTPClient(config *HTTPClientConfig) *http.Client {
	if config == nil {
		config = NewDefaultConfig()
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipTLSVerification,
		},
		MaxIdleConns:    config.MaxIdleConns,
		IdleConnTimeout: config.IdleConnTimeout,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}
}

// CreateHTTPClientWithTLS creates an HTTP client with TLS verification configuration
// Deprecated: Use CreateHTTPClient with HTTPClientConfig instead
func CreateHTTPClientWithTLS(skipTLSVerification bool) *http.Client {
	return CreateHTTPClient(&HTTPClientConfig{
		Timeout:             DefaultTimeout,
		SkipTLSVerification: skipTLSVerification,
		MaxIdleConns:        DefaultMaxIdleConns,
		IdleConnTimeout:     DefaultIdleConnTimeout,
	})
}
