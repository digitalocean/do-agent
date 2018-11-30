package clients

import (
	"net"
	"net/http"
	"time"
)

// HTTPClient is can make HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTP creates a new HTTP client with the provided timeout
func NewHTTP(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
			DisableKeepAlives:     true,
		},
	}
}

// FakeHTTPClient is used for testing
type FakeHTTPClient struct {
	DoFunc func(*http.Request) (*http.Response, error)
}

// Do an HTTP request for testing
func (c *FakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if c.DoFunc != nil {
		return c.DoFunc(req)
	}
	return nil, nil
}
